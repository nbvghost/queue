package queue

import (
	"fmt"
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

	params.Params.MaxWaitCollectMessage = 1000
	params.Params.MaxPoolNum = 10

	p := NewPools()
	//p := NewPools(Order)

	n := 0

	go func() {
		for {

			p.Push(n)

			n++
			if n > 50 {
				// n=0
				//time.Sleep(time.Second*20)
			}
			//time.Sleep(time.Millisecond * 1)

		}
	}()

	go func() {

		for {
			df := p.GetMessage(10)
			fmt.Println(len(df))
			time.Sleep(time.Millisecond * 1000)
		}

	}()

	for {
		p.printStat()
		time.Sleep(time.Second)
	}
}

func TestMemQueue_addIndex(t *testing.T) {
	type fields struct {
		rIndex uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{name: "TestMemQueue_addIndex", fields: fields{rIndex: 1}, want: 2},
		{name: "TestMemQueue_addIndex", fields: fields{rIndex: 10}, want: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MemQueue{
				rIndex: tt.fields.rIndex,
			}
			if got := p.addIndex(); got != tt.want {
				t.Errorf("addIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMemQueue_cutIndex(t *testing.T) {
	type fields struct {
		rIndex uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uint64
	}{
		{name: "TestMemQueue_addIndex", fields: fields{rIndex: 1}, want: 0},
		{name: "TestMemQueue_addIndex", fields: fields{rIndex: 10}, want: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &MemQueue{
				rIndex: tt.fields.rIndex,
			}
			if got := p.cutIndex(); got != tt.want {
				t.Errorf("addIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
