package restrictor

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
