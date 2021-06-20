package RabbitMQ

import (
	"time"

	"github.com/coomp/ccs/log"
)

// GetMQExchange 获取交换机的地址
func GetMQExchange(Serial string) string {
	// 暂时做测试,这里直接返回一个固定地址
	log.L.Debug("get a RabbitMQ exchange SerialId:%s", Serial)
	return "amqp://admin:rabbitmq123@18.232.146.30:5672/"
}

// GetMQTopic 获取topic
func GetMQTopic(Serial string) []string {
	// 这里应该是走配置下发到中控的缓存,当请求来的时候,通过唯一的标识来获取topic
	log.L.Debug("get a RabbitMQ topic SerialId:%s", Serial)
	return []string{"Test_Rmq_topic"}
}

// GetHeartbeatPeriod TODO
func GetHeartbeatPeriod(Serial string) time.Duration {
	// 这里应该是走配置下发到中控的缓存,当请求来的时候,通过唯一的标识来获取topic
	log.L.Debug("get a RabbitMQ GetHeartbeatPeriod SerialId:%s", Serial)
	return time.Duration(20) * time.Second
}
