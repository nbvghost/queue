package queue

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/nbvghost/glog"
	"github.com/nbvghost/queue/block"
	"github.com/nbvghost/queue/params"
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

	params.Params.MaxWaitCollectMessage = 1000
	params.Params.BlockBufferSize = 10
	params.Params.PoolSize = 1000000

	p := NewPools()
	//p := NewPools(Order)

	n := 0

	go func() {
		for {

			err := p.Push(n)
			if err != nil {
				log.Println(err)
				break
			}

			n++
			if n > 500 {

				// n=0
				//time.Sleep(time.Second*20)
			}
			//time.Sleep(time.Millisecond * 1)

		}
	}()

	go func() {

		for {
			p.GetMessage(10)

			//time.Sleep(time.Millisecond * 1000)
		}

	}()

	for {
		p.printStat()
		time.Sleep(time.Second)
	}
}
