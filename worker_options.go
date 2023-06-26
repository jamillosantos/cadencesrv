package srvcadence

type WorkerOption func(o *workerOpts)

type workerOpts struct {
	name string

	domain       string
	taskListName string

	clientName string

	outboundKey string
	hostPort    string

	register []Register
}

func WithName(name string) WorkerOption {
	return func(o *workerOpts) {
		o.name = name
	}
}

func WithDomain(domain string) WorkerOption {
	return func(o *workerOpts) {
		o.domain = domain
	}
}

func WithTaskListName(taskListName string) WorkerOption {
	return func(o *workerOpts) {
		o.taskListName = taskListName
	}
}

func WithClientName(clientName string) WorkerOption {
	return func(o *workerOpts) {
		o.clientName = clientName
	}
}

func WithOutboundKey(outboundKey string) WorkerOption {
	return func(o *workerOpts) {
		o.outboundKey = outboundKey
	}
}

func WithHostPort(hostPort string) WorkerOption {
	return func(o *workerOpts) {
		o.hostPort = hostPort
	}
}

func WithRegister(register ...Register) WorkerOption {
	return func(o *workerOpts) {
		if o.register == nil {
			o.register = register
			return
		}
		o.register = append(o.register, register...)
	}
}
