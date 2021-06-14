package mq

// GetMQExchange 获取交换机的地址
func GetMQExchange() string {
	// 暂时做测试,这里直接返回一个固定地址
	return "amqp://admin:rabbitmq123@18.232.146.30:5672/"
}

// GetRabbitMQTopic 获取topic
func GetRabbitMQTopic() string {
	// 这里应该是走配置下发到中控的缓存,当请求来的时候,通过唯一的标识来获取topic
	return ""
}
