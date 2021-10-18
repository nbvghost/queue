package queue

import (
	"fmt"
	"log"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/queue/block"
)

func init() {

}
func TestPools_Remove(t *testing.T) {
	PoolTool := NewPools()

	type args struct {
		target *block.MemBlock
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "TestPools_Remove", args: args{target: PoolTool.blocks[3]}},
		{name: "TestPools_Remove", args: args{target: PoolTool.blocks[2]}},
		{name: "TestPools_Remove", args: args{target: PoolTool.blocks[0]}},
		{name: "TestPools_Remove", args: args{target: PoolTool.blocks[1]}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//p.Remove(tt.args.target)
		})
	}
	for {
		glog.Trace(PoolTool.blocks)
		time.Sleep(time.Second)
	}

}

func BenchmarkPools_Push(b *testing.B) {
	p := NewPools()
	fmt.Println(b.N)
	for i := 0; i < b.N; i++ {
		p.Push(125)
	}
}
func TestPools_Push(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	//params.Params.MaxWaitCollectMessage = 1000
	//params.Params.BlockBufferSize = 10
	//params.Params.PoolSize = 1000

	p := NewPools()
	//p := NewPools(Order)

	var n int64 = 1

	go func() {
		for {

			for i := 0; i < 10; i++ {
				go func() {
					err := p.Push(atomic.AddInt64(&n, 1))
					if err != nil {
						log.Println(err)
					}

				}()
			}

		}
	}()

	time.Sleep(time.Second)

	go func() {

		for {
			msg := p.GetMessage(1)
			log.Println(msg)
			//time.Sleep(time.Millisecond * 1000)
		}

	}()

	for {
		p.printStat()
		time.Sleep(time.Second)
	}
}
