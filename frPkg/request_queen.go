package frPkg

import (
	"fmt"
	"time"
)

type QueueData struct {
	Func      func() error
	FinalFunc func()
	Title     string
	errNum    int
}

type Priority int

const (
	LowPriority Priority = iota
	MediumPriority
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
	return &Restrictor{
		qps:              qps,
		Chs:              chs,
		onceTime:         time.Second / time.Duration(qps),
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

func (r *Restrictor) nextData() *QueueData {
	chsKeys := []Priority{HighPriority, MediumPriority, LowPriority}
	for _, chsKey := range chsKeys {
		select {
		case chData := <-r.Chs[chsKey]:
			return chData
		default:
		}
	}
	select {
	case chData := <-r.Chs[HighPriority]:
		return chData
	case chData := <-r.Chs[MediumPriority]:
		return chData
	case chData := <-r.Chs[LowPriority]:
		return chData
	case chData := <-r.errCh:
		return chData
	}
}

func (r *Restrictor) runQueenRequest() {
	ticker := time.NewTicker(r.onceTime)
	sem := make(chan struct{}, r.qps)
	go func() {
		for range ticker.C {
			select {
			case sem <- struct{}{}:
			default:
			}
		}
	}()

	for {
		chData := r.nextData()
		<-sem
		go r.doCallBack(chData)
	}
}

func (r *Restrictor) doCallBack(chData *QueueData) {
	err := chData.Func()
	if err != nil {
		go func() {
			if len(r.errCh) >= r.maxErrQueenLen {
				fmt.Println(fmt.Sprintf("[限流器]%s错误队列长度大于%d,%s", r.name, r.maxErrQueenLen, err.Error()))
				chData.FinalFunc()
				return
			}
			if chData.errNum >= r.maxRetryTimes {
				fmt.Println(fmt.Sprintf("[限流器]%s失败次数%d,不再重试,%s", r.name, chData.errNum, err.Error()))
				chData.FinalFunc()
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
