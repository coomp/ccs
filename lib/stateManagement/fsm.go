package stateManagement

import (
	"fmt"
	"strings"
	"sync"
)

// transitioner
type transitioner interface {
	transition(*FSM) error
}

// cKey is a struct key used for keeping the callbacks mapped to a target.
type cKey struct {
	// target is either the name of a state or an event depending on which
	// callback type the key refers to. It can also be "" for a non-targeted
	// callback like before_event.
	target string

	// callbackType is the situation when the callback will be run.
	callbackType int
}

// eKey is a struct key used for storing the transition map.
type eKey struct {
	// event is the name of the event that the keys refers to.
	event string

	// src is the source from where the event can transition.
	src string
}

// FSM TODO
type FSM struct {
	// 当前的状态
	current string

	// transitions 当前状态 + 事件 = 新状态  State(S) + Event(E) => State(S')
	transitions map[eKey]string

	// callbacks 当前状态/当前事件（事件前/事件后/事件中） => 需要回调的方法
	callbacks map[cKey]Callback

	//--------------------------------//
	// transition
	transition func()
	//
	// transitionerObj
	transitionerObj transitioner
	//--------------------------------//

	// stateMu guards access to the current state.
	stateMu sync.RWMutex
	// eventMu guards access to Event() and Transition().
	eventMu sync.Mutex

	// metadata 跨事件存储  SetMetadata() and Metadata() 来做处理
	metadata map[string]interface{}

	metadataMu sync.RWMutex
}

// EventDesc 表示初始化 FSM 时的事件 此FSM支持多种源状态 目的状态只能一个
type EventDesc struct {
	// Name is the event name used when calling for a transition.
	Name string

	// Src is a slice of source states that the FSM must be in to perform a
	// state transition.
	Src []string

	// Dst is the destination state that the FSM will be in if the transition
	// succeds.
	Dst string
}

// Callback 增加上报并且这里需要根据返回结果过规则引擎
type Callback func(*Event)

// Events  NewFSM.
type Events []EventDesc

// Callbacks  callbacks in NewFSM.
type Callbacks map[string]Callback

// NewFSM TODO
func NewFSM(initial string, events []EventDesc, callbacks map[string]Callback) *FSM {
	f := &FSM{
		transitionerObj: &transitionerStruct{},
		current:         initial,                      // 当前状态
		transitions:     make(map[eKey]string),        // 事件映射=>state
		callbacks:       make(map[cKey]Callback),      // 事件回调映射 分为时间前/事件内/事件后
		metadata:        make(map[string]interface{}), // 参数
	}

	// Build transition map and store sets of all events and states.
	allEvents := make(map[string]bool) // 事件集合
	allStates := make(map[string]bool) // 状态集合
	for _, e := range events {
		for _, src := range e.Src {
			f.transitions[eKey{e.Name, src}] = e.Dst //transitions 当前状态 + 事件 = 新状态
			allStates[src] = true
			allStates[e.Dst] = true
		}
		allEvents[e.Name] = true
	}
	// Map all callbacks to events/states.
	for name, fn := range callbacks {
		var target string
		var callbackType int
		switch {
		case strings.HasPrefix(name, "before_"):
			target = strings.TrimPrefix(name, "before_")
			if target == "event" {
				target = ""
				callbackType = callbackBeforeEvent
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackBeforeEvent
			}
		case strings.HasPrefix(name, "leave_"):
			target = strings.TrimPrefix(name, "leave_")
			if target == "state" {
				target = ""
				callbackType = callbackLeaveState
			} else if _, ok := allStates[target]; ok {
				callbackType = callbackLeaveState
			}
		case strings.HasPrefix(name, "enter_"):
			target = strings.TrimPrefix(name, "enter_")
			if target == "state" {
				target = ""
				callbackType = callbackEnterState
			} else if _, ok := allStates[target]; ok {
				callbackType = callbackEnterState
			}
		case strings.HasPrefix(name, "after_"):
			target = strings.TrimPrefix(name, "after_")
			if target == "event" {
				target = ""
				callbackType = callbackAfterEvent
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackAfterEvent
			}
		default:
			target = name
			if _, ok := allStates[target]; ok {
				callbackType = callbackEnterState
			} else if _, ok := allEvents[target]; ok {
				callbackType = callbackAfterEvent
			}
		}

		if callbackType != callbackNone {
			f.callbacks[cKey{target, callbackType}] = fn
		}
		fmt.Printf("name:%s callbackType:%d \n", name, callbackType)
	}

	return f
}

// afterEventCallbacks calls the after_ callbacks, first the named then the
// general version.
func (f *FSM) afterEventCallbacks(e *Event) {
	if fn, ok := f.callbacks[cKey{e.Event, callbackAfterEvent}]; ok {
		fn(e)
	}
	if fn, ok := f.callbacks[cKey{"", callbackAfterEvent}]; ok {
		fn(e)
	}

}

// Current returns the current state of the FSM.
func (f *FSM) Current() string {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return f.current
}

// Is returns true if state is the current state.
func (f *FSM) Is(state string) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	return state == f.current
}

// SetState allows the user to move to the given state from current state.
// The call does not trigger any callbacks, if defined.
func (f *FSM) SetState(state string) {
	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.current = state
	return
}

// Can returns true if event can occur in the current state.
func (f *FSM) Can(event string) bool {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	_, ok := f.transitions[eKey{event, f.current}]
	return ok && (f.transition == nil)
}

// AvailableTransitions returns a list of transitions available in the
// current state.
func (f *FSM) AvailableTransitions() []string {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()
	var transitions []string
	for key := range f.transitions {
		if key.src == f.current {
			transitions = append(transitions, key.event)
		}
	}
	return transitions
}

// Cannot returns true if event can not occure in the current state.
// It is a convenience method to help code read nicely.
func (f *FSM) Cannot(event string) bool {
	return !f.Can(event)
}

// Metadata returns the value stored in metadata
func (f *FSM) Metadata(key string) (interface{}, bool) {
	f.metadataMu.RLock()
	defer f.metadataMu.RUnlock()
	dataElement, ok := f.metadata[key]
	return dataElement, ok
}

// SetMetadata stores the dataValue in metadata indexing it with key
func (f *FSM) SetMetadata(key string, dataValue interface{}) {
	f.metadataMu.Lock()
	defer f.metadataMu.Unlock()
	f.metadata[key] = dataValue
}

// Event initiates a state transition with the named event.
//
// The call takes a variable number of arguments that will be passed to the
// callback, if defined.
//
// It will return nil if the state change is ok or one of these errors:
//
// - event X inappropriate because previous transition did not complete
//
// - event X inappropriate in current state Y
//
// - event X does not exist
//
// - internal error on state transition
//
// The last error should never occur in this situation and is a sign of an
// internal bug.
func (f *FSM) Event(event string, args ...interface{}) error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()

	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if f.transition != nil {
		return fmt.Errorf("InvalidEventError")
	}
	//fmt.Printf("----- transitions:%#v", f.transitions)
	// State(S) + Event(E) =>State(S`) 通过一个初始拿到dst
	dst, ok := f.transitions[eKey{event, f.current}]
	if !ok {
		for ekey := range f.transitions {
			if ekey.event == event {
				return fmt.Errorf("Event And State find a result ,but there is unkowErr")
			}
		}
		return fmt.Errorf("Can not find Event to change sourceState")
	}

	e := &Event{f, event, f.current, dst, nil, args, false, false}

	err := f.beforeEventCallbacks(e)
	if err != nil {
		return err
	}

	if f.current == dst {
		f.afterEventCallbacks(e)
		return fmt.Errorf("afterEventCallbacks")
	}

	// Setup the transition, call it later.
	f.transition = func() {
		f.stateMu.Lock()
		f.current = dst
		f.stateMu.Unlock()

		f.enterStateCallbacks(e)
		f.afterEventCallbacks(e)
	}

	if err = f.leaveStateCallbacks(e); err != nil {
		//if _, ok := err.(CanceledError); ok {
		//	f.transition = nil
		//}
		return err
	}

	// Perform the rest of the transition, if not asynchronous.
	f.stateMu.RUnlock()
	defer f.stateMu.RLock()
	err = f.doTransition()
	if err != nil {
		return fmt.Errorf("doTransition_err:%s", err.Error())
	}

	return e.Err
}

// Transition wraps transitioner.transition.
func (f *FSM) Transition() error {
	f.eventMu.Lock()
	defer f.eventMu.Unlock()
	return f.doTransition()
}

// doTransition wraps transitioner.transition.
func (f *FSM) doTransition() error {
	return f.transitionerObj.transition(f)
}

// transitionerStruct is the default implementation of the transitioner
// interface. Other implementations can be swapped in for testing.
type transitionerStruct struct{}

// Transition completes an asynchrounous state change.
//
// The callback for leave_<STATE> must prviously have called Async on its
// event to have initiated an asynchronous state transition.
func (t transitionerStruct) transition(f *FSM) error {
	if f.transition == nil {
		return fmt.Errorf("transition_nil")
	}
	f.transition()
	f.transition = nil
	return nil
}

// enterStateCallbacks calls the enter_ callbacks, first the named then the
// general version.
func (f *FSM) enterStateCallbacks(e *Event) {
	if fn, ok := f.callbacks[cKey{f.current, callbackEnterState}]; ok {
		fn(e)
	}
	if fn, ok := f.callbacks[cKey{"", callbackEnterState}]; ok {
		fn(e)
	}
}

// beforeEventCallbacks calls the before_ callbacks, first the named then the
// general version.
func (f *FSM) beforeEventCallbacks(e *Event) error {
	if fn, ok := f.callbacks[cKey{e.Event, callbackBeforeEvent}]; ok {
		fn(e)
		if e.canceled {
			return fmt.Errorf("CanceledError_%s", e.Err.Error())
		}
	}
	if fn, ok := f.callbacks[cKey{"", callbackBeforeEvent}]; ok {
		fn(e)
		if e.canceled {
			return fmt.Errorf("CanceledError_%s", e.Err.Error())
		}
	}
	return nil
}

// leaveStateCallbacks calls the leave_ callbacks, first the named then the
// general version.
func (f *FSM) leaveStateCallbacks(e *Event) error {
	if fn, ok := f.callbacks[cKey{f.current, callbackLeaveState}]; ok {
		fn(e)
		if e.canceled {
			return fmt.Errorf("CanceledError_%s", e.Err.Error())
		} else if e.async {
			return fmt.Errorf("AsyncError_%s", e.Err.Error())
		}
	}
	if fn, ok := f.callbacks[cKey{"", callbackLeaveState}]; ok {
		fn(e)
		if e.canceled {
			return fmt.Errorf("CanceledError_%s", e.Err.Error())
		} else if e.async {
			return fmt.Errorf("AsyncError_%s", e.Err.Error())
		}
	}
	return nil
}

const (
	callbackNone int = iota
	callbackBeforeEvent
	callbackLeaveState
	callbackEnterState
	callbackAfterEvent
)
