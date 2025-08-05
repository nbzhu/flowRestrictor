package test

import (
	"fmt"
	"github.com/nbzhu/flowRestrictor/frClient"
	"github.com/nbzhu/flowRestrictor/frPkg"
	"testing"
	"time"
)

func TestOne(t *testing.T) {
	c := frClient.New(100, frPkg.PriorityStruct{HighPriorityLen: 1000, MediumPriorityLen: 1000, LowPriorityLen: 10000})
	c.SetName("测试队列")
	fmt.Printf("开始测试\n")
	for i := 0; i < 10; i++ {
		a := i
		go frClient.ToDo(c, frPkg.LowPriority, frPkg.QueueData{
			Func: func() error {
				fmt.Printf("LowPriority-%d\n", a)
				return nil
			},
			Title: "测试",
		})
	}
	for i := 0; i < 10; i++ {
		a := i
		go frClient.ToDo(c, frPkg.HighPriority, frPkg.QueueData{
			Func: func() error {
				fmt.Printf("HighPriority-%d\n", a)
				return nil
			},
			Title: "测试",
		})
	}
	for i := 0; i < 10; i++ {
		a := i
		go frClient.ToDo(c, frPkg.MediumPriority, frPkg.QueueData{
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
