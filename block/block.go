package block

type Block interface {
	IsFull() bool
	IsEmpty() bool
	TryPush(message interface{}) bool
	Read() <-chan interface{}
}
