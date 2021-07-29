package block

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/queue/params"
)

type MemBlock struct {
	sync.RWMutex
	die         bool
	inputs      []interface{}
	lastInputAt int64
	hash        string
	isFull      bool   // input 是否已经满
	rIndex      uint64 //当前已经读取的index
}

func NewMemBlock() *MemBlock {
	m := &MemBlock{
		inputs: make([]interface{}, 0, params.Params.PoolSize), lastInputAt: time.Now().UnixNano(),
	}
	m.generatorHash()
	return m
}
func (p *MemBlock) generatorHash() string {
	dest := [8]byte{}
	if _, err := rand.Read(dest[:]); err != nil {
		glog.Panic(err)
	}
	p.hash = time.Now().Format("20060102150405.999999999") + "." + hex.EncodeToString(dest[:])
	return p.hash
}

func (p *MemBlock) IsEmpty() bool {
	return len(p.inputs) == 0
}
func (p *MemBlock) Empty() {
	p.Lock()
	defer p.Unlock()
	p.inputs = nil
}
func (p *MemBlock) getIndex() uint64 {
	return atomic.AddUint64(&p.rIndex, 1)
}
func (p *MemBlock) GetNext() (interface{}, error) {
	index := int(p.getIndex())
	if index > len(p.inputs)-1 {
		return nil, errors.New("已经全部读完")
	}
	return p.inputs[index], nil
}
func (p *MemBlock) Push(message interface{}) bool {
	p.Lock()
	defer func() {
		p.Unlock()
		p.lastInputAt = time.Now().UnixNano()
	}()
	if len(p.inputs) >= params.Params.PoolSize {
		p.isFull = false
	} else {
		p.inputs = append(p.inputs, message)
		p.isFull = true
	}
	return p.isFull
}

func (p *MemBlock) IsFull() bool {
	p.RLock()
	defer p.RUnlock()
	return p.isFull
}
