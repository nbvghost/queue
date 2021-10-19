package block

import (
	"sync"
	"sync/atomic"
	"time"
)

type MemBlock struct {
	sync.RWMutex
	lastInputAt int64
	/*
		并发写入,写到 memList 的数量满足 writeNum，将不在写入。todo 没有好的方法将不再写入的  memList 关闭，暂时没有做关闭。
	*/
	memList  chan interface{}
	writeNum int64 //已经写入的数量
	isFull   uint32
	once     *sync.Once
}

func (p *MemBlock) IsFull() bool {
	return atomic.LoadUint32(&p.isFull) == 1
}
func (p *MemBlock) IsEmpty() bool {
	return len(p.memList) == 0
}
func (p *MemBlock) Read() <-chan interface{} {
	return p.memList
}

// TryPush 尝试添加消息，如果满了返回true,使用atomic后，比不用慢了4倍
func (p *MemBlock) TryPush(message interface{}) bool {
	if atomic.AddInt64(&p.writeNum, 1) > int64(cap(p.memList)) {
		if atomic.CompareAndSwapUint32(&p.isFull, 0, 1) {
			close(p.memList)
		}
		return true
	}

	defer func() {
		p.lastInputAt = time.Now().UnixNano()
	}()

	select {
	case p.memList <- message:
		return false
	default:
		return true
	}
}

func NewMemBlock(bufferSize int) Block {
	m := &MemBlock{
		lastInputAt: time.Now().UnixNano(),
		memList:     make(chan interface{}, bufferSize),
		once:        &sync.Once{},
	}
	return m
}
