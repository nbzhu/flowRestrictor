package restrictor

type Base struct {
	qps int
}

type PriorityStruct struct {
	HighPriorityLen   int //高优先级的长度必须设置
	MediumPriorityLen int //若不设置该优先级则不会生成管道
	LowPriorityLen    int //若不设置该优先级则不会生成管道
}

func (b *Base) getRequestChannel(len PriorityStruct) map[Priority]chan *QueueData {
	ma := make(map[Priority]chan *QueueData)
	ma[HighPriority] = make(chan *QueueData, len.HighPriorityLen)
	if len.MediumPriorityLen != 0 {
		ma[MediumPriority] = make(chan *QueueData, len.MediumPriorityLen)
	}
	if len.LowPriorityLen != 0 {
		ma[LowPriority] = make(chan *QueueData, len.LowPriorityLen)
	}
	return ma
}

func (b *Base) SetQps(qps int) {
	b.qps = qps
}
func (b *Base) runQueen(r *Restrictor) {
	r.SetQps(b.qps)
	go r.runQueenRequest()
}

func (b *Base) Run(priorityStruct PriorityStruct) *Restrictor {
	if b.qps == 0 {
		b.qps = 1
	}
	restrictor := NewRestrictor(b.qps, b.getRequestChannel(priorityStruct))
	b.runQueen(restrictor)
	return restrictor
}
