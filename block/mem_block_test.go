package block

import (
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"testing"
)

func TestMemBlock_Push(t *testing.T) {

	t.Run("TestMemBlock_Push", func(t *testing.T) {
		for {
			p := NewMemBlock()

			var ifullCount int64 = 0

			var readCount int64 = 0
			go func() {
				for {
					if msg := p.Read(); msg != nil {
						atomic.AddInt64(&readCount, 1)
					} else {
						//log.Println("readCount",readCount)
						if readCount != 512 {
							panic(errors.New("数据读取不完整"))
						}

					}
				}
			}()

			var wg sync.WaitGroup
			for i := 0; i < 1000; i++ {
				wg.Add(1)
				go func(ii int) {
					defer wg.Done()
					ifull := p.Push(ii)
					if ifull {
						atomic.AddInt64(&ifullCount, 1)

					}

				}(i)
			}
			wg.Wait()

			//if ifullCount != 488 {
			//log.Println(len(p.memList),ifullCount,int64(len(p.memList))+ifullCount)

			//time.Sleep(time.Millisecond*1)
		}
	})
	log.Println("over")
	//time.Sleep(time.Second)

}
