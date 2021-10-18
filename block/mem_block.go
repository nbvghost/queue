package block

import (
	"sync"
	"time"

	"github.com/nbvghost/queue/params"
)

type MemBlock struct {
	sync.RWMutex
	lastInputAt int64
	/*
		并发写入,写到 memList 的数量满足 writeNum，将不在写入。todo 没有好的方法将不再写入的  memList 关闭，暂时没有做关闭。
	*/
	memList  chan interface{}
	writeNum int64 //已经写入的数量
	isFull   bool
	once     *sync.Once
}

func (p *MemBlock) IsEmpty() bool {
	return len(p.memList) == 0
}
func (p *MemBlock) Read() <-chan interface{} {
	return p.memList
}

// TryPush 尝试添加消息，如果满了返回true
func (p *MemBlock) TryPush(message interface{}) bool {
	if p.isFull {
		return p.isFull
	}

	defer func() {
		p.lastInputAt = time.Now().UnixNano()
	}()

	select {
	case p.memList <- message:
		return false
	default:
		p.once.Do(func() {
			close(p.memList)
			p.isFull = true
			p.once = &sync.Once{}
		})
		return true
	}
}
func (p *MemBlock) setFull(v bool) {
	p.Lock()
	defer p.Unlock()
	p.isFull = v
}
func (p *MemBlock) getFull() bool {
	p.RLock()
	defer p.RUnlock()
	return p.isFull
}

func NewMemBlock() *MemBlock {
	m := &MemBlock{
		lastInputAt: time.Now().UnixNano(),
		memList:     make(chan interface{}, params.Params.BlockBufferSize),
		once:        &sync.Once{},
	}
	return m
}
