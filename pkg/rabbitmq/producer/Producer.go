package producer

import (
	"sync"
	"time"

	"git.code.oa.com/tme/hippo-go/service"
	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/pkg/rabbitmq/base"
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
	"github.com/coomp/ccs/pkg/tenant/mq/RabbitMQ"
)

// RabbitMQProducer RabbitMQ 生产者
type RabbitMQProducer struct {
	producerId           string
	clientAddress        string
	groupName            string
	state                comm.NodeState
	heartbeatRetryTimes  int
	lastHeartbeatTime    time.Time
	mux                  *sync.RWMutex
	topics               []string
	topicSelectorMap     map[string]*base.RoundRobinQueueSelector
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
	p.topicSelectorMap = make(map[string]*base.RoundRobinQueueSelector)
	p.mux = new(sync.RWMutex)
	p.heartbeatServiceTick = time.NewTicker(RabbitMQ.GetHeartbeatPeriod(SerialID))
	p.readyQueue = make(chan struct{})
	p.clientAddress = comm.GetLocalIpString()
	p.Init()
	return p, nil
}

// Init TODO
// RabbitMQProducer 初始化
func (p *RabbitMQProducer) Init() {
	var addrPrefix string
	if configs.Conf.Global.Env == 0 {
		addrPrefix = "ip://"
	} else {
		addrPrefix = "dns://"
	}
	rpcclit := clt.NewRpcClient(p.producerConfig.RpcConfig, addrPrefix, p.producerConfig.Master, p.log)
	ms := service.NewMasterService(rpcclit)
	p.masterService = ms
	go p.doHeartBeat(ms)
	go func() {
		for _ = range p.heartbeatServiceTick.C {
			p.doHeartBeat(ms)
		}
	}()
}
