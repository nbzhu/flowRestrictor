package test

import (
	restrictor "FlowRestrictor/pkg/flowRestrictor"
	"fmt"
	"testing"
	"time"
)

func TestOne(t *testing.T) {
	c := restrictor.NewClient(30, restrictor.PriorityStruct{HighPriorityLen: 1000, LowPriorityLen: 10000})

	for i := 0; i < 100; i++ {
		a := i
		restrictor.ToDo(c, restrictor.HighPriority, restrictor.QueueData{
			Func: func() error {
				fmt.Println(a)
				return nil
			},
			Title: "测试",
		})
	}
	time.Sleep(time.Second * 10)
}
