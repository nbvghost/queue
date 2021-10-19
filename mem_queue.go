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

type MemQueue struct {
	blocks     []*block.MemBlock
	PrintTime  time.Time
	maxPoolNum int
	poolNum    int

	inputTotalNum   uint64
	outTotalNum     uint64
	processTotalNum uint64

	locker sync.RWMutex

	pushLocker sync.Mutex

	pushOffset *block.Offset
	readOffset *block.Offset
}

func NewPools() *MemQueue {
	pt := map[string]interface{}{
		"Name":                  "task.Pools",
		"PoolSize":              params.Params.PoolSize,
		"PoolTimeOut":           params.Params.PoolTimeOut,
		"MaxProcessMessageNum":  params.Params.MaxProcessMessageNum,
		"MaxWaitCollectMessage": params.Params.MaxWaitCollectMessage,
		"BlockBufferSize":       params.Params.BlockBufferSize,
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

	if params.Params.BlockBufferSize < 0 {
		panic(errors.New("task.Params.BlockBufferSize,不能为负数"))
	}

	glog.Trace(pt)

	p := &MemQueue{
		pushOffset: &block.Offset{},
		readOffset: &block.Offset{},
		blocks: []*block.MemBlock{
			block.NewMemBlock(params.Params.BlockBufferSize),
		},
	}

	return p

}

func (p *MemQueue) Len() int {

	return len(p.blocks)
}

func (p *MemQueue) GetMessage(num int) []interface{} {
	msgList := make([]interface{}, 0)
	defer func() {
		p.OutMany(uint(len(msgList)))
		p.printStat()
	}()

	if num == 0 {
		num = 1
	}

	if len(p.blocks) == 0 {
		return nil
	}

	for {

		select {
		case <-time.After(time.Duration(params.Params.MaxWaitCollectMessage) * time.Millisecond):
			p.printStat()
			//如果有收集的消息的话，在超时后返回，没有的话，继续收集
			/*if len(msgList) > 0 {
				return msgList
			} else {
				continue
			}*/
			return msgList
		case msg, isOpen := <-p.get():
			if isOpen == false {
				p.cleanAndNext()
			} else {
				if msg != nil {
					msgList = append(msgList, msg)
					if len(msgList) >= num {
						return msgList
					}
				}
			}
			/*default:
			if msg := p.get(); msg != nil {
				msgList = append(msgList, msg)
				if len(msgList) >= num {
					return msgList
				}
			}*/
		}

	}

}

func (p *MemQueue) cleanAndNext() {
	p.locker.Lock()
	defer p.locker.Unlock()
	readIndex := p.readOffset.GetIndex()
	if p.blocks[readIndex].IsEmpty() {
		if readIndex < p.pushOffset.GetIndex() {
			if cuIndex := p.readOffset.Next(); cuIndex > 0 {
				p.blocks = p.blocks[1:]
				p.readOffset.Pre()
				p.pushOffset.Pre()
			}
		}
	}
}

func (p *MemQueue) get() <-chan interface{} {
	item := p.blocks[p.readOffset.GetIndex()].Read()
	return item
}
func (p *MemQueue) getAbleWriteIndex() int {
	if len(p.blocks) == 0 {
		p.scalePool()
	}
	item := p.blocks[len(p.blocks)-1].Read()

	return item
}
func (p *MemQueue) Push(messages ...interface{}) error {
	p.pushLocker.Lock()
	defer p.pushLocker.Unlock()

	for index := range messages {

		for {
			b := p.blocks[len(p.blocks)-1]
			if b.IsFull() {
				if !p.scalePool() {
					return errors.New("pool都满了")
				}
			}
			if b.TryPush(messages[index]) {
				continue
			} else {
				break
			}
		}
	}

	p.InputMany(uint(len(messages)))
	return nil
}
func (p *MemQueue) scalePool() bool {
	p.locker.Lock()
	defer p.locker.Unlock()

	if len(p.blocks) == params.Params.PoolSize {
		return false
	}

	for i := 0; i < 10; i++ {
		if len(p.blocks) <= params.Params.PoolSize {
			p.blocks = append(p.blocks, block.NewMemBlock())
		}
	}
	p.poolNum = len(p.blocks)
	return true
}

func (p *MemQueue) OutMany(num uint64) {
	atomic.AddUint64(&p.outTotalNum, num)
}
func (p *MemQueue) InputMany(num uint64) {
	atomic.AddUint64(&p.inputTotalNum, num)
}
func (p *MemQueue) ProcessOne() {
	atomic.AddUint64(&p.processTotalNum, 1)
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
