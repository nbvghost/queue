package queue

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/queue/block"
	"github.com/nbvghost/queue/params"
)

//type messageOrder string

//const DisOrder messageOrder = "DisOrder"
//const Order messageOrder = "Order"

type MemQueue struct {
	blocks     []*block.MemBlock
	msgChan    chan interface{}
	PrintTime  time.Time
	maxPoolNum int
	poolNum    int

	inputTotalNum   uint
	outTotalNum     uint
	processTotalNum uint

	locker *sync.RWMutex

	totalNumLocker *sync.RWMutex

	once *sync.RWMutex

	workingPool *block.MemBlock

	workingReadPool *block.MemBlock
	rIndex          uint64
}

func (p *MemQueue) addIndex() uint64 {
	return atomic.AddUint64(&p.rIndex, 1)
}
func (p *MemQueue) cutIndex() uint64 {
	return atomic.AddUint64(&p.rIndex, -1)
}
func NewPools() *MemQueue {
	pt := map[string]interface{}{
		"Name":                  "task.Pools",
		"PoolSize":              params.Params.PoolSize,
		"PoolTimeOut":           params.Params.PoolTimeOut,
		"MaxProcessMessageNum":  params.Params.MaxProcessMessageNum,
		"MaxWaitCollectMessage": params.Params.MaxWaitCollectMessage,
		"MaxPoolNum":            params.Params.MaxPoolNum,
	}

	if params.Params.PoolSize <= 0 {
		panic(errors.New("task.Params.PoolSize,不能有零值"))
	}

	if params.Params.PoolTimeOut <= 0 {
		panic(errors.New("task.Params.PoolTimeOut,不能有零值"))
	}

	if params.Params.MaxProcessMessageNum <= 0 {
		panic(errors.New("task.Params.MaxProcessMessageNum,不能有零值"))
	}

	if params.Params.MaxWaitCollectMessage <= 0 {
		panic(errors.New("task.Params.MaxWaitCollectMessage,不能有零值"))
	}

	if params.Params.MaxPoolNum < 0 {
		panic(errors.New("task.Params.MaxPoolNum,不能为负数"))
	}

	glog.Trace(pt)

	p := &MemQueue{msgChan: make(chan interface{}, params.Params.MaxProcessMessageNum),
		totalNumLocker: &sync.RWMutex{},
		once:           &sync.RWMutex{},
		locker:         &sync.RWMutex{}}

	return p

}

func (p *MemQueue) Len() int {

	return len(p.blocks)
}

func (p *MemQueue) scalePool() *block.MemBlock {
	p.locker.Lock()
	defer p.locker.Unlock()

	newPools := make([]*block.MemBlock, 0, 10)
	for i := 0; i < 10; i++ {
		if i < len(p.blocks)-1 {
			if p.blocks[i].IsEmpty() {
				p.blocks = append(p.blocks[:i], p.blocks[i+1:]...)
				p.cutIndex()
			}
		}
		pool := block.NewMemBlock()
		newPools = append(newPools, pool)
	}
	p.blocks = append(p.blocks, newPools...)
	p.poolNum = len(p.blocks)
	return newPools[0]
}

func (p *MemQueue) GetMessage(num int) []interface{} {
	p.once.Lock()
	defer p.once.Unlock()

	msgs := make([]interface{}, 0)
	defer func() {
		p.OutMany(uint(len(msgs)))
		p.printStat()
	}()

	if num == 0 {
		return msgs
	}
	t := time.NewTicker(time.Duration(params.Params.MaxWaitCollectMessage) * time.Millisecond)
	defer t.Stop()

	for {

		select {
		case <-t.C:
			p.printStat()
			//如果有收集的消息的话，在超时后返回，没有的话，继续收集
			if len(msgs) > 0 {
				return msgs
			} else {
				continue
			}
		}

		if msg, err := p.get(); err != nil {

		}

	}

}

func (p *MemQueue) get() (interface{}, error) {
	item, err := p.blocks[p.rIndex].GetNext()
	if err != nil {
		p.addIndex()
		return nil, err
	}
	return item, err
}
func (p *MemQueue) Push(messages ...interface{}) {
	if p.workingPool == nil {
		p.workingPool = p.getAbleWritePool()
	}

	for index := range messages {
		isFull := p.workingPool.Push(messages[index])
		if isFull {
			p.workingPool = p.getAbleWritePool()
			isFull = p.workingPool.Push(messages[index])
			if isFull {
				glog.Debug(fmt.Sprintf("缓冲区已满"))
			}
		}
	}

	p.InputMany(uint(len(messages)))
}

func (p *MemQueue) getAbleWritePool() *block.MemBlock {
	var ablePool *block.MemBlock
	p.locker.Lock()
	defer p.locker.Unlock()
	length := len(p.blocks)
	if length > 0 {
		_ablePool := p.blocks[length-1]
		if _ablePool.IsFull() == false {
			ablePool = _ablePool
		}
	}
	if ablePool == nil {
		ablePool = p.scalePool()
	}
	return ablePool
}

func (p *MemQueue) OutMany(num uint) {
	p.totalNumLocker.Lock()
	defer p.totalNumLocker.Unlock()
	p.outTotalNum = p.outTotalNum + num
}
func (p *MemQueue) InputMany(num uint) {
	p.totalNumLocker.Lock()
	defer p.totalNumLocker.Unlock()
	p.inputTotalNum = p.inputTotalNum + num
}
func (p *MemQueue) ProcessOne() {
	p.totalNumLocker.Lock()
	defer p.totalNumLocker.Unlock()
	p.processTotalNum++
}
func (p *MemQueue) IsEmpty() bool {

	for i := 0; i < len(p.blocks); i++ {
		if p.blocks[i].IsEmpty() == false {
			return false
		}
	}

	p.printStat()
	if p.outTotalNum != p.inputTotalNum || p.outTotalNum != p.processTotalNum {
		return false
	}

	return true

}
func (p *MemQueue) printStat() {
	now := time.Now()
	if now.Sub(p.PrintTime) > time.Second*10 {
		p.poolNum = len(p.blocks)
		glog.Trace(fmt.Sprintf("MaxPoolNum:%v   PoolNum:%v  InputTotalNum:%v   OutTotalNum:%v   ProcessTotalNum:%v", p.maxPoolNum, p.poolNum, p.inputTotalNum, p.outTotalNum, p.processTotalNum))
		p.PrintTime = now
	}
}
