package limap

import (
	"context"
	"golang.org/x/time/rate"
	"sync"
	"time"

	"github.com/twmb/murmur3"
)

type Limap struct {
	lock             sync.Mutex
	Expiry           time.Duration
	SweepingInterval time.Duration
	Rate             int
	Burst            int
	m                map[uint64]limapEntry
	StopBGSweeper    context.CancelFunc
}

type limapEntry struct {
	rateLimiter *rate.Limiter
	lastSeen    time.Time
}

func New(rate, burst int, expiry time.Duration, initSize int, sweepingInterval time.Duration) (m *Limap) {
	if rate < 1 {
		rate = 10
	}
	if burst < 1 {
		burst = 0
	}
	if expiry < 1 {
		expiry = time.Hour
	}
	if sweepingInterval < 1 {
		sweepingInterval = time.Second
	}

	m = &Limap{}
	m.Rate = rate
	m.Burst = burst
	m.Expiry = expiry
	m.SweepingInterval = sweepingInterval
	m.m = make(map[uint64]limapEntry, initSize)

	ctx, cancel := context.WithCancel(context.Background())
	m.StopBGSweeper = cancel
	go m.RunBGSweeper(ctx)

	return
}

func (m *Limap) Set(key []byte) {
	k := murmur3.Sum64(key)

	m.lock.Lock()

	fe := limapEntry{}
	fe.lastSeen = time.Now()
	fe.rateLimiter = rate.NewLimiter(rate.Limit(m.Rate), m.Burst)

	m.m[k] = fe

	m.lock.Unlock()
}

func (m *Limap) Del(key []byte) {
	k := murmur3.Sum64(key)

	m.lock.Lock()

	delete(m.m, k)

	m.lock.Unlock()
}

func (m *Limap) IsAllowed(key []byte) (allowed, ok bool) {
	k := murmur3.Sum64(key)

	var fe limapEntry
	fe, ok = m.m[k]
	if !ok {
		return
	}

	m.lock.Lock()

	fe.lastSeen = time.Now()
	m.m[k] = fe

	m.lock.Unlock()

	allowed = m.m[k].rateLimiter.Allow()

	return
}

func (m *Limap) RunBGSweeper(ctx context.Context) {
	t := time.NewTicker(m.SweepingInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
		default:
		}

		m.lock.Lock()

		for k, v := range m.m {
			if time.Since(v.lastSeen) >= m.Expiry {
				delete(m.m, k)
			}
		}

		m.lock.Unlock()

		<-t.C
	}
}
