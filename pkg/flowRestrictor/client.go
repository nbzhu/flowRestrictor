package restrictor

import "errors"

type Client struct {
	*Base
}

func NewClient(qps int, priority PriorityStruct) *Restrictor {
	client := &Client{Base: &Base{}}
	client.SetQps(qps)
	return client.Run(priority)
}

func ToDo(Queen *Restrictor, p Priority, queueData QueueData) {
	Queen.Chs[p] <- &queueData
}

func TryToDo(queen *Restrictor, p Priority, queueData QueueData) error {
	select {
	case queen.Chs[p] <- &queueData:
		return nil
	default:
		return errors.New("当前队列已满")
	}
}
