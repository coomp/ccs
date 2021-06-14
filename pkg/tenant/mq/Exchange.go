package mq

// Exchange 交换机信息
type Exchange struct {
	name         string            // 交换器名称
	exchangeType string            // 交换器类型，有如下四种，direct，topic，fanout，headers
	durability   bool              // 是否需要持久化，true为持久化。持久化可以将交换器存盘，在服务器重启的时候不会丢失相关信息
	autoDelete   bool              // 与这个Exchange绑定的Queue或Exchange都与此解绑时，会删除本交换器
	internal     bool              // 设置是否内置，true为内置。如果是内置交换器，客户端无法发送消息到这个交换器中，只能通过交换器路由到交换器这种方式
	argument     map[string]string // 其他一些结构化参数
}

// SetName 设置 交换器名称
func (e *Exchange) SetName(n string) {
	e.name = n
}

// SetExchangeType 设置 消息的编码类型
func (e *Exchange) SetExchangeType(exchangeType string) {
	e.exchangeType = exchangeType
}

// SetDurability 设置 持久化
func (e *Exchange) SetDurability(d bool) {
	e.durability = d
}

// SetAutoDelete 设置 是否删除交换机
func (e *Exchange) SetAutoDelete(a bool) {
	e.autoDelete = a
}

// SetInternal 设置 消息的持久化类型
func (e *Exchange) SetInternal(i bool) {
	e.internal = i
}

// SetArgument 设置 用户自定义任意的键和值
func (e *Exchange) SetArgument(a map[string]string) {
	e.argument = a
}
