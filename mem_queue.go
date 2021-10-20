package queue

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/queue/block"
)

type MemQueue struct {
	blocks     []block.Block
	PrintTime  time.Time
	maxPoolNum int
	poolNum    int

	inputTotalNum   uint64
	outTotalNum     uint64
	processTotalNum uint64

	locker sync.RWMutex

	config Config
}

func NewQueue(config Config) *MemQueue {
	pt := map[string]interface{}{
		"Name":                  "task.Pools",
		"PoolSize":              config.PoolSize,
		"PoolTimeOut":           config.PoolTimeOut,
		"MaxProcessMessageNum":  config.MaxProcessMessageNum,
		"MaxWaitCollectMessage": config.MaxWaitCollectMessage,
		"BlockBufferSize":       config.BlockBufferSize,
	}

	if config.PoolSize <= 0 {
		config.PoolSize = 1024
	}

	if config.PoolTimeOut <= 0 {
		config.PoolTimeOut = 60 * 1000
	}

	if config.MaxProcessMessageNum <= 0 {
		config.MaxProcessMessageNum = 1000
	}

	if config.MaxWaitCollectMessage <= 0 {
		config.MaxWaitCollectMessage = 1000
	}
	if config.NextWaitReadMessage <= 0 {
		config.NextWaitReadMessage = 800
	}
	if config.BlockBufferSize <= 0 {
		config.BlockBufferSize = 512
	}

	glog.Trace(pt)

	p := &MemQueue{
		config: config,
		blocks: []block.Block{
			block.NewMemBlock(config.BlockBufferSize),
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
		p.OutMany(uint64(len(msgList)))
		p.printStat()
	}()

	if num == 0 {
		num = 1
	}
	if num > p.config.MaxProcessMessageNum {
		num = p.config.MaxProcessMessageNum
	}

	if len(p.blocks) == 0 {
		return nil
	}

	for {

		select {
		case <-time.After(time.Duration(p.config.MaxWaitCollectMessage) * time.Millisecond):
			p.printStat()
			return msgList
		case msg, isOpen := <-p.blocks[0].Read():
			if isOpen == false {
				p.removeBlock()
			} else {
				if msg != nil {
					msgList = append(msgList, msg)
					if len(msgList) >= num {
						return msgList
					}
				}
			}
		}

	}

}

func (p *MemQueue) removeBlock() {
	p.locker.Lock()
	defer p.locker.Unlock()

	if len(p.blocks) > 0 {
		if p.blocks[0].IsFull() && p.blocks[0].IsEmpty() {
			p.blocks = p.blocks[1:]
		}
	}

}
func (p *MemQueue) Push(messages ...interface{}) {
	p.locker.Lock()
	defer p.locker.Unlock()

	defer func() {
		p.InputMany(uint64(len(messages)))
	}()
	for index := range messages {
		for {
			b := p.getLatest()
			if b.IsFull() {
				b = p.scalePool()
			}
			if b.TryPush(messages[index]) {
				continue
			} else {
				break
			}
		}
	}
}
func (p *MemQueue) getLatest() block.Block {
	//p.locker.RLock()
	//defer p.locker.RUnlock()
	b := p.blocks[len(p.blocks)-1]
	return b
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

func (p *MemQueue) scalePool() block.Block {

	if len(p.blocks) >= p.config.PoolSize {
		p.blocks = append(p.blocks, block.NewDiskBlock())
	} else {
		p.blocks = append(p.blocks, block.NewMemBlock(p.config.BlockBufferSize))
	}
	p.poolNum = len(p.blocks)
	return p.blocks[len(p.blocks)-1]
}

func (p *MemQueue) printStat() {
	now := time.Now()
	if now.Sub(p.PrintTime) > time.Second*10 {
		p.poolNum = len(p.blocks)
		glog.Trace(fmt.Sprintf("MaxPoolNum:%v   PoolNum:%v  InputTotalNum:%v   OutTotalNum:%v   ProcessTotalNum:%v", p.maxPoolNum, p.poolNum, p.inputTotalNum, p.outTotalNum, p.processTotalNum))
		p.PrintTime = now
	}
}
