package producer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.code.oa.com/going/attr"
	"git.code.oa.com/tme/hippo-go/monitor"
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
	// 从租户分或者从配置,暂时都从配置
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

// SendMessageWithTimout TODO
func (p *RabbitMQProducer) SendMessageWithTimout(msg *comm.Message, du time.Duration) (*SendResult, error) {
	if e := comm.EncodeHeader(); e != nil {
		return nil, e
	}
	//if !p.checkState(du) {
	//	return nil, errors.New("state error")
	//}
	//if e := p.checkMessage(msg); e != nil {
	//	return nil, e
	//}
	topic := msg.Topic
	selector := p.topicSelectorMap[topic]
	if selector == nil {
			≥1) //[HippoProducer]当前topic的RoundRobinQueueSelector为空
		return nil, errors.New(
			"ERROR-1-TDBANK-Hippo|00003|ERROR|NO_AVAILABLE_QUEUE|There is no available selector for topic " + topic + ",maybe you pass a wrong topic name")
	}
	queue := selector.Select()
	if queue == nil {
		return nil, errors.New("ERROR-1-TDBANK-Hippo|00003|ERROR|NO_AVAILABLE_QUEUE|There is no available queue for topic " +
			topic + ",maybe you don't publish it at first or the topic temporarily unavailable")
	}
	svc := p.getBrokerService(&queue.BrokerGroup.Brokers[0])
	res, err := svc.SendMessage(p.createSendMessageRequest(msg, queue.Id))
	if err != nil {
		return nil, err
	}
	if res.Success {
		msg.Id = res.MessageId
		return &SendResult{Success: true, Code: res.Code, Queue: queue, Message: msg}, nil
	}
	if res.Code != def.NOT_MASTER {
		return &SendResult{Success: false, Code: res.Code, Queue: queue, ErrorMsg: res.Tips}, nil
	}
	p.rebalanceTopic(topic, queue.Id, res.BrokerGroup)
	svc = p.getBrokerService(&queue.BrokerGroup.Brokers[0])
	res, err = svc.SendMessage(p.createSendMessageRequest(msg, queue.Id))
	if err != nil {
		attr.AttrAPI(monitor.HIPPO_PRODUCER_SEND_MSG_SECOND_FAIL, 1) //[HippoProducer]sendMessage-second失败
		return nil, err
	}
	if res.Success {
		msg.Id = res.MessageId
		return &SendResult{Success: true, Code: res.Code, Queue: queue, Message: msg}, nil
	}
	return &SendResult{Success: false, Code: res.Code, Queue: queue, ErrorMsg: res.Tips}, nil

}

// SendMessage TODO
func (p *RabbitMQProducer) SendMessage(msg *def.Message) (*SendResult, error) {
	return p.SendMessageWithTimout(msg, 6*time.Second)
}

// ProduceMessage TODO
func (p *RabbitMQProducer) ProduceMessage(topic string, data []byte) error {
	m := &def.Message{Data: data, Topic: topic}
	res, err := p.SendMessage(m)
	if err != nil {
		return err
	}
	if res != nil && res.Code != 0 {
		return errors.New(fmt.Sprintf("SendMsgFail the res.code=%d res.errmsg=%s", res.Code, res.ErrorMsg))
	}
	return nil
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
