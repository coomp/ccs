package base

// Queue
type Queue struct {
	_msgpack       struct{}        `msgpack:",asArray"`
	Class          string          `json:"@type,omitempty" msgpack:"-"`
	Id             int64           `json:"id,omitempty"`
	ExchangeGroup  *ExchangeGroups `json:"exchangeGroups,omitempty"`
	Topic          string          `json:"topic,omitempty"`
	Model          string          `json:"model,omitempty"`
	Writable       bool            `json:"writable,omitempty"`
	Readable       bool            `json:"readable,omitempty"`
	BrokerGroupKey int32           `json:"brokerGroupKey,omitempty"`
	Status         int32           `json:"status,omitempty"`
}

type RoundRobinQueueSelector struct {
	ql  []Queue
	act uint64
}

