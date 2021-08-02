package block

import "sync"

type Offset struct {
	index  int64
	locker sync.RWMutex
}

func (p *Offset) GetIndex() int64 {
	p.locker.RLock()
	defer p.locker.RUnlock()
	return p.index
}
func (p *Offset) Next() int64 {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.index++
	return p.index
}
func (p *Offset) Pre() int64 {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.index--
	if p.index < 0 {
		p.index = 0
	}
	return p.index
}
