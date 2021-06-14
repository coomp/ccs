package comm

type NodeState int

const (
	READY   NodeState = 0
	RUNNING NodeState = 1
	STOP    NodeState = 2
)
