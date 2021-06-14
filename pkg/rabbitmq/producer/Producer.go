package producer

import (
	"coomp/log"
	"coomp/pkg/rabbitmq/comm"
	"coomp/pkg/tenant/mq/RabbitMQ"
	"sync"
	"time"
)

// RabbitMQProducer RabbitMQ 生产者
type RabbitMQProducer struct {
	producerId           string
	clientAddress        string
	groupName            string
	producerConfig       *config.ProducerConfig
	state                def.NodeState
	heartbeatRetryTimes  int
	lastHeartbeatTime    time.Time
	masterService        *service.MasterService
	brokerServices       map[string]*service.BrokerWriteService
	mux                  *sync.RWMutex
	topics               []string
	topicSelectorMap     map[string]*RoundRobinQueueSelector
	heartbeatServiceTick *time.Ticker
	readyQueue           chan struct{}
}

// NewRabbitMQProducer NewRabbitMQProducer 创建
func NewRabbitMQProducer(SerialID string) (*RabbitMQProducer, error) {


	p := new(RabbitMQProducer)
	p.state = comm.READY
	p.topics = RabbitMQ.GetMQTopic(SerialID)
	p.groupName = RabbitMQ.GetMQExchange(SerialID)
	p.producerId = comm.GenerateMQID(p.groupName)
	p.heartbeatRetryTimes = 0
	p.lastHeartbeatTime = time.Now()
	p.topicSelectorMap = make(map[string]*RoundRobinQueueSelector)
	p.brokerServices = make(map[string]*service.BrokerWriteService)
	p.mux = new(sync.RWMutex)
	p.heartbeatServiceTick = time.NewTicker(sec.HeartbeatPeriod)
	p.readyQueue = make(chan struct{})
	p.clientAddress = comm.GetLocalIpString()
	p.Init()
	return p, nil
}

// RabbitMQProducer 初始化
func (p *RabbitMQProducer) Init() {
	var addrPrefix string
	if p.producerConfig.IsTest {
		addrPrefix = "ip://"
	} else {
		addrPrefix = "dns://"
	}
	rpcclit := clt.NewHippoRpcClient(p.producerConfig.RpcConfig, addrPrefix, p.producerConfig.Master, p.log)
	ms := service.NewMasterService(rpcclit)
	p.masterService = ms
	go p.doHeartBeat(ms)
	go func() {
		for _ = range p.heartbeatServiceTick.C {
			p.doHeartBeat(ms)
		}
	}()
}

