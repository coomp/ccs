package fsm

// Event is the info that get passed as a reference in the callbacks.
type Event struct {
	// FSM is a reference to the current FSM.
	FSM *FSM
	// Event is the event name.
	Event string
	// Args is a optinal list of arguments passed to the callback.
	Args []interface{}

	// cmd http/tcp/udp
	Cmd string

	// Address  ip:port
	Address string
}

// 这里写通用回调
// beforeEventCallbacks 主要看当前上一个流程状态的回调有没有返回正确,这里注入规则引擎
// afterEventCallbacks 下一个流程的调用,拿到结果给下一个流程的beforeEventCallbacks
