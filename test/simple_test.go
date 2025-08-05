package test

import (
	restrictor "FlowRestrictor/pkg/flowRestrictor"
	"fmt"
	"testing"
	"time"
)

func TestOne(t *testing.T) {
	c := restrictor.NewClient(100, restrictor.PriorityStruct{HighPriorityLen: 1000, MediumPriorityLen: 1000, LowPriorityLen: 10000})
	c.SetName("测试队列")
	fmt.Printf("开始测试\n")
	for i := 0; i < 10; i++ {
		a := i
		go restrictor.ToDo(c, restrictor.LowPriority, restrictor.QueueData{
			Func: func() error {
				fmt.Printf("LowPriority-%d\n", a)
				return nil
			},
			Title: "测试",
		})
	}
	for i := 0; i < 10; i++ {
		a := i
		go restrictor.ToDo(c, restrictor.HighPriority, restrictor.QueueData{
			Func: func() error {
				fmt.Printf("HighPriority-%d\n", a)
				return nil
			},
			Title: "测试",
		})
	}
	for i := 0; i < 10; i++ {
		a := i
		go restrictor.ToDo(c, restrictor.MediumPriority, restrictor.QueueData{
			Func: func() error {
				fmt.Printf("MediumPriority-%d\n", a)
				return nil
			},
			Title: "测试",
		})
	}

	time.Sleep(time.Second * 3)
	fmt.Printf("over")
}
