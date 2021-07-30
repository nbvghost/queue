package queue

import (
	"errors"
	"fmt"
	"sync"
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

	locker         sync.RWMutex
	totalNumLocker sync.RWMutex

	blockPushIndex       int64
	blockPushIndexLocker sync.RWMutex

	//blockReadIndex       int64
	//blockReadIndexLocker sync.RWMutex
}

//push
func (p *MemQueue) getPushBlockIndex() int64 {
	p.blockPushIndexLocker.RLock()
	defer p.blockPushIndexLocker.RUnlock()
	return p.blockPushIndex
}
func (p *MemQueue) nextPushBlockIndex() int64 {
	p.blockPushIndexLocker.Lock()
	defer p.blockPushIndexLocker.Unlock()
	p.blockPushIndex++
	return p.blockPushIndex
}
func (p *MemQueue) prePushBlockIndex() int64 {
	p.blockPushIndexLocker.Lock()
	defer p.blockPushIndexLocker.Unlock()
	p.blockPushIndex--
	return p.blockPushIndex
}

//read
/*func (p *MemQueue) getReadBlockIndex() int64 {
	p.blockReadIndexLocker.RLock()
	defer p.blockReadIndexLocker.RUnlock()
	return p.blockReadIndex
}
func (p *MemQueue) nextReadBlockIndex() int64 {
	p.blockReadIndexLocker.Lock()
	defer p.blockReadIndexLocker.Unlock()
	p.blockReadIndex++
	return p.blockReadIndex
}
func (p *MemQueue) preReadBlockIndex() int64 {
	p.blockReadIndexLocker.Lock()
	defer p.blockReadIndexLocker.Unlock()
	p.blockReadIndex--
	return p.blockReadIndex
}*/

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

	p := &MemQueue{msgChan: make(chan interface{}, params.Params.MaxProcessMessageNum)}

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

	for {

		select {
		case <-time.After(time.Duration(params.Params.MaxWaitCollectMessage) * time.Millisecond):
			p.printStat()
			//如果有收集的消息的话，在超时后返回，没有的话，继续收集
			if len(msgList) > 0 {
				return msgList
			} else {
				continue
			}
		default:
			if msg := p.get(); msg != nil {
				msgList = append(msgList, msg)
				if len(msgList) >= num {
					return msgList
				}
			}
		}

	}

}

func (p *MemQueue) get() interface{} {
	if len(p.blocks) == 0 {
		return nil
	}
	//log.Println(l,p.blockIndex)
	item := p.blocks[0].Read()
	if item == nil {
		p.locker.Lock()
		defer p.locker.Unlock()
		if len(p.blocks) > 1 {
			p.blocks[0].Free()
			p.blocks = p.blocks[1:]
		}
		return nil
	}
	return item
}
func (p *MemQueue) Push(messages ...interface{}) error {
	for index := range messages {
		for {
			pushIndex := p.getPushBlockIndex()
			if int(pushIndex) > len(p.blocks)-1 {
				if p.scalePool() == false {

					return errors.New("扩容失败")
				}
				continue
			}
			isFull := p.blocks[pushIndex].Push(messages[index])
			if isFull {
				p.nextPushBlockIndex()
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
		if len(p.blocks) < params.Params.PoolSize {
			p.blocks = append(p.blocks, block.NewMemBlock())
		}
	}
	p.poolNum = len(p.blocks)
	return true
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
