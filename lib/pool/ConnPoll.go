package pool

import (
	"container/list"
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

	chInitialized uint32        // set to 1 when field ch is initialized 原子锁ch初始化一次
	mu            sync.Mutex    // 锁
	closed        bool          // set to true when the pool is closed.
	Active        int           // 连接池中打开的连接数
	ch            chan struct{} // limits open connections when p.Wait is true
	Idle          list.List     // idle 连接
}

// ThePoolConn 连接的双向链表结构体
type thePoolConn struct {
	C       net.Conn
	h       time.Time // 心跳声时间
	created time.Time //创建时间
}

// NewPool create a pool with capacity
func NewPool(maxCap int, newFunc func() (net.Conn, error)) (*Pool, error) {
	if maxCap == 0 {
		return nil, fmt.Errorf("invalid capacity settings")
	}
	p := new(Pool)
	if newFunc != nil {
		p.Dial = newFunc
	}
	// 默认初始化一个,维持的有更大空闲连接，这里不必纠结
	for i := 0; i < 1; i++ {
		v, err := p.create()
		if err != nil {
			return p, err
		}
		e := &thePoolConn{
			C:       v,
			h:       time.Now(),
			created: time.Now(),
		}
		// 头尾是一个
		p.Idle.PushBack(e)
		p.MaxActive = maxCap
		p.MaxIdle++
	}
	return p, nil
}

// RegisterChecker 定时清理垃圾
func (p *Pool) RegisterChecker(interval time.Duration, check func(interface{}) bool) {
	if interval > 0 && check != nil {
		go func() {
			for {
				time.Sleep(interval)
				p.mu.Lock()
				if p.Idle.Len() == 0 {
					// pool aleardy destroyed, exit
					p.mu.Unlock()
					return
				}
				// 新的不检查
				p.mu.Unlock()
				for node := p.Idle.Front(); node != p.Idle.Back(); node = node.Next() {
					if (interval*10 < time.Now().Sub(node.Value.(thePoolConn).h) && p.MaxActive-p.Active > p.MaxIdle) || p.IdleTimeout*
						time.Second > time.Now().Sub(node.Value.(thePoolConn).created) || p.TestOnBorrow(node.Value.(thePoolConn).C,
						p.MaxConnLifetime) {
						// 上一次心跳的时间到当前已经超过两个周期了,说明该清理了,但是具体清理不清理掉,要看是不是最小连接数,还有是否有效
						// 首先看是否有效,无效就删除掉 这个地方是不是应该大一点，2个心跳周期太短了，3分钟？嗯，10个周期吧
						// 超过超时时间也不要了重新创建吧
						p.mu.Lock()
						node.Value.(thePoolConn).C.Close()
						p.Idle.Remove(node)
						p.MaxIdle--
						p.mu.Unlock()
					}
				}
			}
		}()
	}
}

func (p *Pool) create() (net.Conn, error) {
	if p.Dial == nil {
		return nil, fmt.Errorf("Pool.New is nil, can not create connection")
	}
	return p.Dial()
}

// Len returns current connections in pool
func (p *Pool) Len() int {
	return p.Idle.Len()
}

// Get TODO
func (p *Pool) Get() (*thePoolConn, error) {
	// 空的时候创建
	if p.Idle.Len() == 0 {
		c, err := p.create()
		if err != nil {
			return nil, err
		}
		poolConnFront := &thePoolConn{
			C:       c,
			h:       time.Now(),
			created: time.Now(),
		}
		// 只有空的时候才会初始化一个,所以头尾是一样的
		p.mu.Lock()
		p.Idle.PushBack(poolConnFront)
		// 正常情况下这里不应该会进来,说明很久么有连接了，所有的连接都超时被t了，才会过来，这里应该会有一个默认值 就是newpool的时候的值
		// p.MaxActive = 1000 MaxIdle 应该也是0,这里创建出一个，立马又被使用了
		// p.MaxIdle++ // 因为创建所以++
		// p.MaxIdle-- // 被使用了所以-- ,所以注释了上面几行，但是要知道这个运行情况
		p.Active++ // 因为是get 所以活跃的+1
		p.mu.Unlock()
		return poolConnFront, nil
	}
	// 剩下的直接拿，拿不到就等 同时 check下心跳 再同时 检查下是否还有效
	// 当然先判定一下是否已经最大链接数了
	if !p.Wait && p.MaxActive > 0 && p.Active >= p.MaxActive {
		p.mu.Unlock()
		return nil, errors.New("pool 耗尽了")
	}
	p.mu.Lock()
	// idle的连接，都有效 因为会有异步程序来做清理，当然不能肯定
	//从Idle list 获取一个头部的空闲链接，因为比较早创建
	pc := p.Idle.Remove(p.Idle.Front())
	p.Active++  // 忙碌+1
	p.MaxIdle-- // 空闲-1
	p.mu.Unlock()
	return &thePoolConn{
		C:       pc.(thePoolConn).C,
		h:       pc.(thePoolConn).h,
		created: pc.(thePoolConn).created,
	}, nil
}
