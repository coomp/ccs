package app

import (
	"fmt"
	"time"

	"github.com/coomp/ccs/configs"
	"github.com/coomp/ccs/lib/fsm"
	"github.com/coomp/ccs/log"
)

// FSMApp TODO
type FSMApp struct {
}

// FSMList TODO
var FSMList []*fsm.FSM

type event struct {
	Name      string
	Src       string
	Dst       string
	Callbacks map[string]string //Note: If it is not marked whether it is pre-event or post-event, it will be called by default pre-event
}

type fsmInfo struct {
	init      string
	eventList []event
}

// Config TODO
type Config struct {
	fsms []fsmInfo
}

func init() {
	// 10s 变更一次 ,服务启动15s后开始服务,避开初次加载的时候fsm空的现象
	t := time.NewTimer(time.Second * 2)
	defer t.Stop()

	go func() {
		for {
			<-t.C
			if configs.Cfg != nil {
				// 如果不是空的,初始化进来
				FsmConfList := getFsmConf(configs.Cfg)
				fmt.Println(FsmConfList)
				//fmt.Println("fmt.Println(configList)", configs.Cfg)
			}
			// need reset
			t.Reset(time.Second * 2)
		}
	}()
}

func getEvent(conf map[interface{}]interface{}) (event event) {
	event.Callbacks = make(map[string]string)
	if _, ok := conf["Callbacks"]; ok {
		for k, v := range conf["Callbacks"].(map[interface{}]interface{}) {
			event.Callbacks[k.(string)] = v.(string)
		}
	}
	if _, ok := conf["Dst"]; ok {
		event.Dst = conf["Dst"].(string)
	}
	if _, ok := conf["Dst"]; ok {
		event.Src = conf["Dst"].(string)
	}
	if _, ok := conf["Name"]; ok {
		event.Name = conf["Dst"].(string)
	}
	return
}

func getEvenList(conf map[interface{}]interface{}) (fsmInfo fsmInfo) {
	if _, ok := conf["evenList"]; ok {
		for _, v := range conf["evenList"].([]interface{}) {
			e := getEvent(v.(map[interface{}]interface{}))
			fsmInfo.eventList = append(fsmInfo.eventList, e)
		}
	}
	return
}

func getInit(conf map[interface{}]interface{}) (fsmInfo fsmInfo) {
	if _, ok := conf["init"]; ok {
		fsmInfo.init = conf["init"].(string)
	}
	return
}

func getfsm(conf map[interface{}]interface{}) (fsmInfo fsmInfo) {
	fsmInfo.init = getInit(conf).init
	fsmInfo.eventList = getEvenList(conf).eventList
	return
}

/*
map[fsm:map[evenList:[map[Callbacks:map[before_scan:callback_before] Dst:scanning Name:scan Src:idle] map[Callbacks:map[before_scan:callback_working] Dst:scanning Name:working Src:scanning] map[Callbacks:map[before_scan:callback_situation] Dst:scanning Name:situation Src:scanning] map[Callbacks:map[before_scan:callback_working] Dst:scanning Name:working Src:scanning]] init:idle]]
*/

func getFsmConf(config interface{}) (configList Config) {
	var conf = make(map[interface{}]interface{})
	for _, v := range config.([]interface{}) {
		conf = v.(map[interface{}]interface{})
		if _, ok := conf["fsm"]; ok {
			//将 map 转换为指定的结构体
			configList.fsms = append(configList.fsms, getfsm(conf["fsm"].(map[interface{}]interface{})))
		}
	}
	return

}

// Run TODO
func (app *FSMApp) Run() {
	log.L.Debug("there is log test")
	fsm := fsm.NewFSM(
		"idle",
		fsm.Events{
			{Name: "scan", Src: []string{"idle"}, Dst: "scanning"},
			{Name: "working", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"idle"}, Dst: "idle"},
			{Name: "finish", Src: []string{"scanning"}, Dst: "idle"},
		},
		fsm.Callbacks{
			"before_scan": func(e *fsm.Event) {
				fmt.Println("1111after_scan: " + e.FSM.Current())
				fmt.Println("after_scan: " + e.FSM.Current())
			},
			"working": func(e *fsm.Event) {
				fmt.Println("working: " + e.FSM.Current())
			},
			"situation": func(e *fsm.Event) {
				fmt.Println("situation: " + e.FSM.Current())
			},
			"finish": func(e *fsm.Event) {
				fmt.Println("finish: " + e.FSM.Current())
			},
		},
	)

	fmt.Println(fsm.Current())

	err := fsm.Event("scan")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("1:" + fsm.Current())

	err = fsm.Event("working")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("2:" + fsm.Current())

	err = fsm.Event("situation")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("3:" + fsm.Current())

	err = fsm.Event("finish")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("4:" + fsm.Current())
}
