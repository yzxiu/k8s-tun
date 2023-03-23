package counter

import "sync/atomic"

var TotalReadByte uint64 = 0

var TotalWriteByte uint64 = 0

func IncrReadByte(n int) {
	atomic.AddUint64(&TotalReadByte, uint64(n))
}

func IncrWriteByte(n int) {
	atomic.AddUint64(&TotalWriteByte, uint64(n))
}
