package wire

import (
	"math"
	"sync/atomic"
)

// sequence 序列号
type sequence struct {
	num uint32
}

// Next 返回下一个序列号
func (s *sequence) Next() uint32 {
	next := atomic.AddUint32(&s.num, 1)
	if next == math.MaxUint32 {
		if atomic.CompareAndSwapUint32(&s.num, next, 1) {
			return 1
		}
		return s.Next()
	}
	return next
}

// Seq 初始序列号
var Seq = sequence{num: 1}
