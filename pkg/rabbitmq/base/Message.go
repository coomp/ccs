package base

// Message 消息的设置
type Message struct {
	contentType     string            // 消息体的MIME类型，如application/json
	contentEncoding string            // 消息的编码类型，如是否压缩
	messageId       string            // 消息的唯一性标识，由应用进行设置
	timestamp       int64             // 消息的创建时刻，整型，精确到秒
	deliveryMode    int               // 消息的持久化类型 ，1为非持久化，2为持久化，性能影响巨大
	headers         map[string]string // 键/值对表，用户自定义任意的键和值
	priority        int               // 指定队列中消息的优先级
}

// SetContentType 设置 消息体的MIME类型
func (m *Message) SetContentType(t string) {
	m.contentType = t
}

// SetContentEncoding 设置 消息的编码类型
func (m *Message) SetContentEncoding(e string) {
	m.contentEncoding = e
}

// SetContentType 设置 消息的唯一性标识
func (m *Message) SetMessageId(messageId string) {
	m.messageId = messageId
}

// SetContentType 设置 消息体的MIME类型
func (m *Message) SetTimestamp(t int64) {
	m.timestamp = t
}

// SetContentType 设置 消息的持久化类型
func (m *Message) SetDeliveryMode(d int) {
	m.deliveryMode = d
}

// SetContentType 设置 用户自定义任意的键和值
func (m *Message) SetHeaders(h map[string]string) {
	m.headers = h
}

// SetContentType 设置 用户自定义任意的键和值
func (m *Message) SetPriority(p int) {
	m.priority = p
}
