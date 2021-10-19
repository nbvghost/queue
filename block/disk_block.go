package block

type DiskBlock struct {
}

func (d *DiskBlock) IsFull() bool {
	panic("implement me")
}

func (d *DiskBlock) IsEmpty() bool {
	panic("implement me")
}

func (d *DiskBlock) TryPush(message interface{}) bool {
	panic("implement me")
}

func (d *DiskBlock) Read() <-chan interface{} {
	panic("implement me")
}

func NewDiskBlock() Block {
	return &DiskBlock{}
}
