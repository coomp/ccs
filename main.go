package main

import (
	"fmt"
	"unsafe"

	"github.com/coomp/ccs/lib/stateManagement"
	"github.com/coomp/ccs/log"
)

// RealTimeLabelInfo TODO
type RealTimeLabelInfo struct {
	HistoricalPayment int
	NewUser           int
	ExperienceUser    int
	OpeningUser       int
	LstOpenType       int
	LstPaymentAmount  int
	LstSource         int
	LstPaymentMethods int
}

// main 工程入口
func main() {
	var info RealTimeLabelInfo
	info.LstOpenType = 1
	info.LstPaymentMethods = 1
	info.LstSource = 2
	fmt.Println(unsafe.Sizeof(info))
	//producer, err := producer.NewRabbitMQProducer("")
	//producer.Init()
	log.L.Debug("there is log test")
	fsm := stateManagement.NewFSM(
		"idle",
		stateManagement.Events{
			{Name: "scan", Src: []string{"idle"}, Dst: "scanning"},
			{Name: "working", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"scanning"}, Dst: "scanning"},
			{Name: "situation", Src: []string{"idle"}, Dst: "idle"},
			{Name: "finish", Src: []string{"scanning"}, Dst: "idle"},
		},
		stateManagement.Callbacks{
			"before_scan": func(e *stateManagement.Event) {
				fmt.Println("1111after_scan: " + e.FSM.Current())
				fmt.Println("after_scan: " + e.FSM.Current())
			},
			"working": func(e *stateManagement.Event) {
				fmt.Println("working: " + e.FSM.Current())
			},
			"situation": func(e *stateManagement.Event) {
				fmt.Println("situation: " + e.FSM.Current())
			},
			"finish": func(e *stateManagement.Event) {
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
