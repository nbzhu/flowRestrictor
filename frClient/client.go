package frClient

import (
	"errors"
	"github.com/nbzhu/flowRestrictor/frPkg"
)

type Client struct {
	*frPkg.Base
}

func New(qps int, priority frPkg.PriorityStruct) *frPkg.Restrictor {
	client := &Client{Base: &frPkg.Base{}}
	client.SetQps(qps)
	return client.Run(priority)
}

func ToDo(Queen *frPkg.Restrictor, p frPkg.Priority, queueData frPkg.QueueData) {
	Queen.Chs[p] <- &queueData
}

func TryToDo(queen *frPkg.Restrictor, p frPkg.Priority, queueData frPkg.QueueData) error {
	select {
	case queen.Chs[p] <- &queueData:
		return nil
	default:
		return errors.New("当前队列已满")
	}
}
