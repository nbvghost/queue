package block

import (
	"log"
	"testing"
	"time"
)

func BenchmarkMemBlock_TryPush(b *testing.B) {
	p := NewMemBlock(b.N)
	for i := 0; i < b.N; i++ {
		isFull := p.TryPush(i)
		if isFull {

		}
	}
}
func TestMemBlock_Push(t *testing.T) {
	t.Run("TestMemBlock_Push", func(t *testing.T) {

		p := NewMemBlock(10)

		go func() {
			for msg := range p.Read() {
				log.Println(msg)
			}
			log.Println("read over 1")
		}()
		go func() {
			for msg := range p.Read() {
				log.Println(msg)
			}
			log.Println("read over 2")
		}()

		//var num int64

		for {
			for n := 0; n < 200; n++ {
				go func() {
					wp := n //atomic.AddInt64(&num, 1)
					if !p.IsFull() {
						if p.TryPush(wp) {
							log.Println("is full", wp)
						}
					} else {
						log.Println("is full", wp)
					}
				}()

			}
		}
	})
	log.Println("over")
	time.Sleep(time.Second * 30)

}
