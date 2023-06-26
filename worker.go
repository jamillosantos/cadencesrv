package srvcadence

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jamillosantos/logctx"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
)

type Register interface {
	Register(r worker.Registry)
}

type Worker struct {
	ctx context.Context

	name string

	clientName   string
	domain       string
	taskListName string

	register []Register

	outboundKey string
	hostPort    string

	waitChM sync.Mutex
	waitCh  chan struct{}
}

func NewWorker(ctx context.Context, opts ...WorkerOption) *Worker {
	o := defaultWorkerOpts()
	for _, opt := range opts {
		opt(&o)
	}
	w := &Worker{
		ctx:          ctx,
		name:         o.name,
		clientName:   o.clientName,
		domain:       o.domain,
		taskListName: o.taskListName,
		register:     o.register,
		outboundKey:  o.outboundKey,
		hostPort:     o.hostPort,
	}
	return w
}

var (
	outboundIndex atomic.Int32
)

func defaultWorkerOpts() workerOpts {
	return workerOpts{
		name:        "Worker",
		outboundKey: fmt.Sprintf("cadence-%d", outboundIndex.Add(1)),
	}
}

func (w *Worker) Name() string {
	return w.name
}

func (w *Worker) Listen(_ context.Context) error {
	workerOptions := worker.Options{
		Logger: logctx.From(w.ctx),
	}

	cadenceWorker := worker.New(
		w.buildCadenceClient(),
		w.domain,
		w.taskListName,
		workerOptions,
	)

	w.waitChM.Lock()
	w.waitCh = make(chan struct{})
	w.waitChM.Unlock()

	for _, r := range w.register {
		r.Register(cadenceWorker)
	}

	go func() {
		select {
		case <-w.ctx.Done():
		case <-w.waitCh:
		}
		cadenceWorker.Stop()
	}()

	return cadenceWorker.Start()
}

func (w *Worker) buildCadenceClient() workflowserviceclient.Interface {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(w.clientName))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: w.clientName,
		Outbounds: yarpc.Outbounds{
			w.outboundKey: {Unary: ch.NewSingleOutbound(w.hostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(w.outboundKey))
}

func (w *Worker) Close(ctx context.Context) error {
	w.waitChM.Lock()
	close(w.waitCh)
	w.waitChM.Unlock()
	return nil
}
