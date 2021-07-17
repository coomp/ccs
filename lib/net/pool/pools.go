package pool

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Pool TODO
type Pool struct {
	// 建立tcp连接
	Dial func() (net.Conn, error)

	// 健康检测，判断连接是否断开
	TestOnBorrow func(c net.Conn, t time.Duration) bool

	// 连接池中最大空闲连接数
	MaxIdle int

	// 打开最大的连接数
	MaxActive int

	// Idle多久断开连接，小于服务器超时时间
	IdleTimeout time.Duration

	// 配置最大连接数的时候，并且wait是true的时候，超过最大的连接，get的时候会阻塞，直到有连接放回到连接池
	Wait bool

	// 超过多久时间 链接关闭
	MaxConnLifetime time.Duration

	chInitialized uint32     // set to 1 when field ch is initialized 原子锁ch初始化一次
	mu            sync.Mutex // 锁
	closed        bool       // set to true when the pool is closed.
	// Active        int           // 连接池中打开的连接数
	ch   chan struct{} // limits open connections when p.Wait is true
	Idle idleList      // idle 连接
}

// 空闲连，记录poolConn的头和尾
type idleList struct {
	count       int
	front, back *poolConn
}

// 连接的双向链表结构体
type poolConn struct {
	C          net.Conn
	heartbeat  time.Time
	created    time.Time
	next, prev *poolConn
}

// 添加头结点
func (l *idleList) pushFront(pc *poolConn) {
	pc.next = l.front
	pc.prev = nil
	if l.count == 0 {
		l.back = pc
	} else {
		l.front.prev = pc
	}
	l.front = pc
	l.count++
	return
}

// 释放头结点
func (l *idleList) popFront() {
	pc := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.next.prev = nil
		l.front = pc.next
	}
	pc.next, pc.prev = nil, nil
}

// 释放尾结点
func (l *idleList) popBack() {
	pc := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		pc.prev.next = nil
		l.back = pc.prev
	}
	pc.next, pc.prev = nil, nil
}

// NewPool create a pool with capacity
func NewPool(maxCap, MaxActive int, newFunc func() (net.Conn, error)) (*Pool, error) {
	if maxCap == 0 || MaxActive == 0 {
		return nil, fmt.Errorf("invalid capacity settings")
	}
	p := new(Pool)
	if newFunc != nil {
		p.Dial = newFunc
	} else {
		return nil, fmt.Errorf("invalid newFunc")
	}
	// 默认初始化一个,维持的有更大空闲连接，这里不必纠结
	for i := 0; i < 1; i++ {
		v, err := p.create()
		if err != nil {
			return p, err
		}

		// 头尾是一个
		p.Idle = v
		p.MaxIdle = maxCap
		p.MaxActive = MaxActive
	}
	return p, nil
}

// RegisterChecker 定时清理垃圾 这个函数长了一些 不过还是没有超过100行,等后面看看拆解下
func (p *Pool) RegisterChecker(interval time.Duration, check func(interface{}) bool) {
	if interval > 0 && check != nil {
		go func() {
			for {
				time.Sleep(interval)
				p.mu.Lock()
				if p.Idle.count == 0 {
					// pool aleardy destroyed, exit
					p.mu.Unlock()
					return
				}
				// 拿出尾节点
				pc := p.Idle.back
				for i := 0; i < p.Idle.count; i++ {
					// 检查了超时/网络不可用/是否超过最大连接数 TODO 这里没有维持一个最小连接数的逻辑,需要做一个？
					if p.IdleTimeout > 0 && time.Now().Sub(pc.heartbeat) > p.IdleTimeout {
						// 超时了就应该干掉
						if p.Idle.count > 1 {
							pc.C.Close()
							// 后节点的前指针指向当前节点的前节点
							pc.next.prev = pc.prev
							// 前节点的后指针指向当前节点的后节点
							pc.prev.next = pc.next
							p.Idle.count-- // 存在的连接数-1
							//p.Active--     // 活跃的链接数-1
						} else {
							// 此时一定是1，释放该节点 原地切换所以不用加减  因为是++操作，不会无限循环这里
							pc.C.Close()     // 唯一的一个关闭掉
							p.Idle.popBack() // 从双向链表中去掉
							// 创建一个新节点
							v, err := p.create() // 创建一个新的
							if err != nil {
								return
							}
							// 头尾是一个
							p.Idle = v // 放进去
						}

					} else {
						// 这里进行清理,只保留最小连接数 说明都在超时范围内
						if p.MaxActive < p.Idle.count {
							// 需要删减
							pc.C.Close()
							// 后节点的前指针指向当前节点的前节点
							pc.next.prev = pc.prev
							// 前节点的后指针指向当前节点的后节点
							pc.prev.next = pc.next
							p.Idle.count-- // 存在的连接数-1
						}
					}

					// 上面超时时间没有问题的话，就要看下是否网络上仍然是ok的
					if !check(pc) {
						if p.Idle.count > 1 {
							pc.C.Close()
							// 后节点的前指针指向当前节点的前节点
							pc.next.prev = pc.prev
							// 前节点的后指针指向当前节点的后节点
							pc.prev.next = pc.next
							p.Idle.count-- // 存在的连接数-1
						} else {
							// 此时一定是1，释放该节点 原地切换所以不用加减  因为是++操作，不会无限循环这里
							pc.C.Close()     // 唯一的一个关闭掉
							p.Idle.popBack() // 从双向链表中去掉
							// 创建一个新节点
							v, err := p.create() // 创建一个新的
							if err != nil {
								return
							}
							// 头尾是一个
							p.Idle = v // 放进去
						}
						continue
					}
					pc = pc.prev //向前滑动
				}
				p.mu.Unlock()
			}
		}()
	}
}

func (p *Pool) create() (idleList, error) {
	conn, err := p.Dial()
	if err != nil {
		return idleList{}, err
	}
	e := &poolConn{
		C:         conn,
		heartbeat: time.Now(),
		created:   time.Now(),
	}
	idle := &idleList{}
	idle.pushFront(e)
	return *idle, nil
}

// Len returns current connections in pool
func (p *Pool) Len() int {
	return p.Idle.count
}

// Get TODO
func (p *Pool) Get() (*poolConn, error) {
	// 空的时候创建
	if p.Idle.count == 0 {
		// 一般不会进入这里，除非有bug,连接池创建好之后有任务不断的去check,当剩下最后一个时候就会维持至少一个
		c, err := p.create()
		if err != nil {
			return nil, err
		}
		p.mu.Lock()
		p.Idle = c // 放进去
		p.mu.Unlock()
		return p.Idle.back, nil
	}
	// 剩下的直接拿，拿不到就等 同时 check下心跳 再同时 检查下是否还有效
	// 当然先判定一下是否已经最大链接数了
	if p.Wait && p.MaxActive > 0 && p.Idle.count >= p.MaxActive {
		return nil, errors.New("pool 耗尽了")
	}
	p.mu.Lock()
	// 从尾部拿
	c := p.Idle.front
	// 把头部的信息去掉，防止重拿
	p.Idle.popBack()
	p.Idle.count-- // 剩余--
	p.mu.Unlock()
	return c, nil
}

// Put set back conn into store again
func (p *Pool) Put(v interface{}) {
	// 无脑放入,通过check来做排异，从头部放
	c := v.(poolConn)
	p.Idle.pushFront(&c)
	p.Idle.count++ // 剩余++
}
