package queue

import (
	"fmt"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

func init() {

}

func BenchmarkPools_Push(b *testing.B) {
	p := NewQueue(Config{})
	fmt.Println(b.N)
	for i := 0; i < b.N; i++ {
		p.Push(i)
	}
}
func TestPools_Push(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	//params.Params.MaxWaitCollectMessage = 1000
	//params.Params.BlockBufferSize = 10
	//params.Params.PoolSize = 1000

	p := NewQueue(Config{})
	//p := NewPools(Order)

	go func() {
		for n := 0; n < 1000; n++ {
			for i := 0; i < 100; i++ {
				go func(ii int) {
					p.Push(ii)
				}((n * 1000 * 100) + i)
			}
		}
	}()

	time.Sleep(time.Second)

	go func() {

		for {
			time.Sleep(time.Second * 10)
			p.Push(19999999)
		}

	}()

	go func() {

		var pp int64
		go func() {
			for {
				log.Println(pp)
				time.Sleep(time.Second)
			}
		}()
		for {
			msg := p.GetMessage(1)
			log.Println(msg)
			//time.Sleep(time.Millisecond * 1000)
			atomic.AddInt64(&pp, 1)
		}

	}()

	for {
		p.printStat()
		time.Sleep(time.Second)
	}
}
