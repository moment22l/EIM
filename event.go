package EIM

import (
	"sync"
	"sync/atomic"
)

// Event 表示一个会在未来某时刻出现的一次性event
type Event struct {
	fired int32
	c     chan struct{}
	o     sync.Once
}

// Fire 使e完成。同时调用多次是安全的。如果对Fire的调用导致Done返回的信令信道关闭，则返回true。
func (e *Event) Fire() bool {
	ret := false
	e.o.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.c)
		ret = true
	})
	return ret
}

// Done 返回一个调用Fire后会被关闭的chan
func (e *Event) Done() <-chan struct{} {
	return e.c
}

// HasFired 如果e已经调用过Fire则返回true
func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

// NewEvent 返回一个新的等待使用的Event
func NewEvent() *Event {
	return &Event{c: make(chan struct{})}
}
