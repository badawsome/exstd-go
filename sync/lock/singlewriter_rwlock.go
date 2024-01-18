package rwlock

import (
	"runtime"
	"sync"

	pid "github.com/choleraehyq/pid"
)

const (
	cacheLineSize = 64
)

var (
	shardsLen int
)

type SingleWriterRWLock []rwmutexShard

type rwmutexShard struct {
	_ [cacheLineSize]byte
	sync.RWMutex
}

func InitSingleWriterRWLockRwLock() {
	shardsLen = runtime.GOMAXPROCS(0)
}

func NewSingleWriterRWLock() SingleWriterRWLock {
	return SingleWriterRWLock(make([]rwmutexShard, shardsLen))
}

func (lock SingleWriterRWLock) Lock() {
	for shard := range lock {
		lock[shard].Lock()
	}
}

func (lock SingleWriterRWLock) Unlock() {
	for shard := range lock {
		lock[shard].Unlock()
	}
}

func (lock SingleWriterRWLock) RLocker() sync.Locker {
	tid := pid.GetPid()
	return lock[tid%shardsLen].RWMutex.RLocker()
}
