package block

type DiskBlock struct {
}

func (d *DiskBlock) IsFull() bool {
	return false
}

func (d *DiskBlock) IsEmpty() bool {
	return false
}

func (d *DiskBlock) TryPush(message interface{}) bool {
	return false
}

func (d *DiskBlock) Read() <-chan interface{} {
	return make(chan interface{}, 10)
}

func NewDiskBlock() Block {

	return &DiskBlock{}
}
