package producer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/def"
	"github.com/coomp/ccs/log"
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
	clt                  *base.MasterService
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
	err := p.Init()
	return p, err
}

// Init TODO
// RabbitMQProducer 初始化
func (p *RabbitMQProducer) Init() error {
	// TODO 从租户分配,从配置读取
	s := strings.Split(configs.Conf.RpcConfig.DataSourceName, "?")
	if len(s) < 2 {
		return fmt.Errorf("init err,case config is err")
	}

	mc := base.NewMasterService(base.NewRabbitmqRpcClient(&configs.Conf.RpcConfig, s[0]))
	go p.doHeartBeat(mc)
	go func() {
		for range p.heartbeatServiceTick.C {
			p.doHeartBeat(mc)
		}
	}()
	return nil
}

func (p *RabbitMQProducer) createHeartBeatRequest() *comm.HeartbeatRequest {
	req := &comm.HeartbeatRequest{
		ClientId: p.producerId,
		Group:    p.groupName,
		Topics:   p.topics,
		Version:  1,
	}
	return req
}

func (p *RabbitMQProducer) doHeartBeat(clt *base.MasterService) {
	if p.state == def.STOP {
		return
	}
	rs, err := clt.Hearbeat(p.createHeartBeatRequest())
	if err != nil {
		log.L.Errorf("heartbeat error %s\n", err.Error())
	}
	if !rs.Success || rs.Tips != "" {
		log.L.Errorf("hearbeat not success %v,%s\n", rs.Success, rs.Tips)
		p.heartbeatRetryTimes++
		return
	}
	if rs.Code == 201 {
		p.lastHeartbeatTime = time.Now()
		return
	}

	shareBroker := rs.ShareBrokerGroupMap
	topicQueue := rs.TopicQueues
	if len(shareBroker) != 0 && len(topicQueue) != 0 {
		for k, v := range topicQueue {
			for i := range v {
				v[i].Topic = k
				temp := shareBroker[strconv.Itoa(int(v[i].BrokerGroupKey))]
				v[i].BrokerGroup = &temp
			}
		}
	}
}
