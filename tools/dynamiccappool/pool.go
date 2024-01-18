package dynamiccappool

import (
	"sync/atomic"
	"time"

	"github.com/badawsome/exstd-go/sync/lock"
)

const recapInterval = time.Second

type Pool struct {
	semaphore semaphore
	mu        lock.SingleWriterRWLock
	capTodo   atomic.Int32

	close chan interface{}
}

type semaphore struct {
	cap  atomic.Int32
	size atomic.Int32
}

func (sem *semaphore) Resize(cap int, size int) {
	sem.cap.Store(int32(cap))
	sem.cap.Store(int32(size))
}

func NewPool(cap int) *Pool {
	pool := &Pool{
		semaphore: semaphore{},
		capTodo:   atomic.Int32{},
		mu:        lock.NewSingleWriterRWLock(),
	}
	pool.semaphore.Resize(cap, cap)
	pool.capTodo.Store(int32(cap))
	go pool.resize()
	return pool
}

func (p *Pool) resize() {
	ticker := time.NewTicker(recapInterval)
	for {
		select {
		case <-ticker.C:
			capNew := p.capTodo.Load()
			capOld := p.semaphore.cap.Load()
			if capNew != capOld {
				p.mu.Lock()
				p.semaphore.cap.Store(capNew)
				sizeOld := p.semaphore.size.Load()
				hasAcquired := capOld - sizeOld
				p.semaphore.size.Store(capNew - hasAcquired)
				p.mu.Unlock()
			}
		case <-p.close:
			return
		}
	}
}

func (p *Pool) Close() {
	close(p.close)
}

func (p *Pool) Acquire() bool {
	p.mu.RLocker().Lock()
	defer p.mu.RLocker().Unlock()
	if now := p.semaphore.cap.Load(); now > 0 {
		return p.semaphore.cap.CompareAndSwap(now, now-1)
	}
	return false
}

func (p *Pool) Release() {
	p.mu.RLocker().Lock()
	defer p.mu.RLocker().Unlock()
	p.semaphore.cap.Add(1)
}

func (p *Pool) Recap(cap int) {
	p.capTodo.Store(int32(cap))
}
