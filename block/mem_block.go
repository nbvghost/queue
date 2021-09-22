package block

import (
	"log"
	"sync"
	"sync/atomic"
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
}

func NewMemBlock() *MemBlock {
	m := &MemBlock{
		lastInputAt: time.Now().UnixNano(),
		memList:     make(chan interface{}, params.Params.BlockBufferSize),
	}
	return m
}
func (p *MemBlock) IsEmpty() bool {
	return len(p.memList) == 0
}
func (p *MemBlock) Free() {
	//close(p.memList)
}
func (p *MemBlock) Read() <-chan interface{} {
	/*select {
	case msg := <-p.memList:
		return msg
	default:
		return nil
	}*/
	return p.memList
}

// Push 添加超时
func (p *MemBlock) Push(message interface{}) bool {
	if p.getFull() {
		return true
	}
	defer func() {
		p.lastInputAt = time.Now().UnixNano()
	}()

	if num := atomic.AddInt64(&p.writeNum, 1); num <= int64(cap(p.memList)) { //对 p.writeNum 进行累加，返回的num是无序的
		select {
		case p.memList <- message:
			log.Println(num)
			/*if atomic.LoadInt64(&p.writeNum)==int64(cap(p.memList)){
				log.Println("dd")
			}*/
			//log.Println(num, atomic.LoadInt64(&p.writeNum), int64(cap(p.memList)), atomic.LoadInt64(&p.writeNum) == int64(cap(p.memList)))
			return false
		default:

			return true
		}
	} else {
		p.setFull(true)
		n := atomic.AddInt64(&p.writeNum, -1)
		if n == int64(cap(p.memList)) {

			//close(p.memList) //在这里关闭的话，会提前把channel关了
		}
	}
	return true
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
