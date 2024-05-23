package restrictor

import (
	"fmt"
	"time"
)

type QueueData struct {
	Func   func() error
	Title  string
	errNum int
}

type Priority int

const (
	LowPriority Priority = iota
	HighPriority
)

type Restrictor struct {
	name             string
	qps              int
	onceTime         time.Duration
	Chs              map[Priority]chan *QueueData
	errCh            chan *QueueData
	maxErrQueenLen   int
	maxRetryTimes    int
	noticeRetryTimes int
}

func NewRestrictor(qps int, chs map[Priority]chan *QueueData) *Restrictor {
	millisecond := 1000 / qps
	microsecond := int((float32(1000)/float32(qps) - float32(millisecond)) * 1000)
	return &Restrictor{
		qps:              qps,
		Chs:              chs,
		onceTime:         time.Duration(int64(millisecond))*time.Millisecond + time.Duration(int64(microsecond))*time.Microsecond,
		errCh:            make(chan *QueueData, 9999),
		maxErrQueenLen:   9000,
		maxRetryTimes:    5,
		noticeRetryTimes: 2,
	}
}

func (r *Restrictor) SetName(name string) {
	r.name = name
}
func (r *Restrictor) SetMaxErrQueenLen(len int) *Restrictor {
	r.maxErrQueenLen = len
	return r
}
func (r *Restrictor) SetMaxRetryTimes(num int) *Restrictor {
	r.maxRetryTimes = num
	return r
}
func (r *Restrictor) SetNoticeRetryTimes(num int) *Restrictor {
	r.noticeRetryTimes = num
	return r
}

func (r *Restrictor) SetQps(qps int) {
	r.qps = qps
}
func (r *Restrictor) runQueenRequest() {
	for {
		select {
		case chData := <-r.Chs[HighPriority]:
			go r.doCallBack(chData)
		case chData := <-r.Chs[LowPriority]:
		loop:
			for {
				select {
				case chDataInner := <-r.Chs[HighPriority]:
					go r.doCallBack(chDataInner)
				default:
					break loop
				}
				time.Sleep(r.onceTime)
			}
			go r.doCallBack(chData)
		case chData := <-r.errCh:
			go r.doCallBack(chData)
		}

		time.Sleep(r.onceTime)
	}
}

func (r *Restrictor) doCallBack(chData *QueueData) {
	err := chData.Func()
	if err != nil {
		go func() {
			if len(r.errCh) >= r.maxErrQueenLen {
				fmt.Println(fmt.Sprintf("[限流器]%s错误队列长度大于%d,%s", r.name, r.maxErrQueenLen, err.Error()))
				return
			}
			if chData.errNum >= r.maxRetryTimes {
				fmt.Println(fmt.Sprintf("[限流器]%s失败次数%d,不再重试,%s", r.name, chData.errNum, err.Error()))
				return
			}
			if chData.errNum >= r.noticeRetryTimes {
				fmt.Println(fmt.Sprintf("[限流器]%s失败次数%d,%s", r.name, chData.errNum, err.Error()))
			}
			if chData.errNum >= 3 {
				time.Sleep(time.Second * 8)
			}
			time.Sleep(time.Second * 9)
			chData.errNum++
			r.errCh <- chData
		}()
		return
	}
}
