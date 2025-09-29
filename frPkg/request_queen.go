package frPkg

import (
	"context"
	"log"
	"time"
)

type QueueData struct {
	Ctx       context.Context
	Func      func() error
	FinalFunc func(err error)
	Title     string
	errNum    int
}

type Priority int

const (
	LowPriority Priority = iota
	MediumPriority
	HighPriority
)

type RestrictorType uint8

const (
	RestrictorTypeTokenBucket   RestrictorType = iota //令牌桶
	RestrictorTypeSlidingWindow                       //滑动窗口
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
	restrictorType   RestrictorType
	reqTimestamps    []int64
	printCtxErr      bool
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
		reqTimestamps:    make([]int64, 0, qps+10),
	}
}

func (r *Restrictor) SetRestrictorType(t RestrictorType) *Restrictor {
	r.restrictorType = t
	return r
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
func (r *Restrictor) SetPrintCtxErr() *Restrictor {
	r.printCtxErr = true
	return r
}

func (r *Restrictor) setQps(qps int) {
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
	switch r.restrictorType {
	case RestrictorTypeSlidingWindow:
		r.runSlidingWindowQueenRequest()
	case RestrictorTypeTokenBucket:
		r.runTokenBucketQueenRequest()
	default:
		r.runTokenBucketQueenRequest()
	}
}

func (r *Restrictor) runTokenBucketQueenRequest() {
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
		if r.skipByCtxDone(chData) {
			continue
		}
		<-sem
		go r.doCallBack(chData)
	}
}

func (r *Restrictor) skipByCtxDone(chData *QueueData) bool {
	ctx := chData.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		if r.printCtxErr {
			log.Printf("%s skip chData: ctx done: %v", chData.Title, err)
		}
		go chData.FinalFunc(ctx.Err())
		return true
	}
	return false
}

func (r *Restrictor) slidingWindowAllow() bool {
	now := time.Now().UnixMilli()
	cutoff := now - 1000

	i := 0
	for i < len(r.reqTimestamps) && r.reqTimestamps[i] <= cutoff {
		i++
	}
	if i > 0 {
		r.reqTimestamps = r.reqTimestamps[i:]
	}

	if len(r.reqTimestamps) < r.qps {
		r.reqTimestamps = append(r.reqTimestamps, now)
		return true
	}
	return false
}

func (r *Restrictor) runSlidingWindowQueenRequest() {
	for {
		chData := r.nextData()
		if r.skipByCtxDone(chData) {
			continue
		}
		for {
			if r.slidingWindowAllow() {
				go r.doCallBack(chData)
				break
			}
			// 等待短时间再试（可根据需要调整）
			time.Sleep(5 * time.Millisecond)
		}
	}
}

func (r *Restrictor) doCallBack(chData *QueueData) {
	err := chData.Func()
	if err != nil {
		if len(r.errCh) >= r.maxErrQueenLen {
			log.Printf("[限流器]%s错误队列长度大于%d,%s", r.name, r.maxErrQueenLen, err.Error())
			if chData.FinalFunc != nil {
				chData.FinalFunc(err)
			}
			return
		}
		if chData.errNum >= r.maxRetryTimes {
			log.Printf("[限流器]%s失败次数%d,不再重试,%s", r.name, chData.errNum, err.Error())
			if chData.FinalFunc != nil {
				chData.FinalFunc(err)
			}
			return
		}
		if chData.errNum >= r.noticeRetryTimes {
			log.Printf("[限流器]%s失败次数%d,%s", r.name, chData.errNum, err.Error())
		}
		if chData.errNum >= 3 {
			time.Sleep(time.Second * 8)
		}
		time.Sleep(time.Second * 9)
		chData.errNum++
		r.errCh <- chData
	} else {
		if chData.FinalFunc != nil {
			chData.FinalFunc(err)
		}
	}
}
