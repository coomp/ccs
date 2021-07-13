package stateManagement

import (
	"fmt"
)

// Turnstile 闸机，状态机就是和闸机一样的功能，单一状态机的实现.一个业务一个状态机,单业务定制,状态机不维护请求对象的状态，仅仅做逻辑
type Turnstile struct {
	ID     uint64   // 状态机的名字
	State  string   // 当前状态
	States []string // 状态流转按顺序
}

// OnEnter 进入闸机
func (p *Turnstile) OnEnter(toState string, args []interface{}) error {
	if len(args) == 0 || (args[0].(*Turnstile)).State != "create" {
		return fmt.Errorf("OnExit|there is  args err")
	}

	t := args[0].(*Turnstile)
	t.State = toState
	t.States = append(t.States, toState)

	return nil
}

// OnExit TODO
func (p *Turnstile) OnExit(fromState string, args []interface{}) {

}
