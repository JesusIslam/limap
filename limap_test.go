package limap_test

import (
	"testing"
	"time"

	"github.com/JesusIslam/limap"
	"github.com/stretchr/testify/require"
)

func TestLimap(t *testing.T) {
	var l *limap.Limap

	t.Run("When new default", func(t *testing.T) {
		l = limap.New(0, 0, 0, 0, 0)
		require.NotNil(t, l)
	})

	t.Run("When new valid", func(t *testing.T) {
		rate := 2
		burst := 1
		expiry := 2 * time.Second
		initSize := 10
		sweepingInterval := time.Second

		l = limap.New(rate, burst, expiry, initSize, sweepingInterval)
		require.NotNil(t, l)
	})

	key := []byte("192.168.1.1")

	t.Run("When set ok", func(t *testing.T) {
		l.Set(key)

		allowed, ok := l.IsAllowed(key)
		require.True(t, allowed)
		require.True(t, ok)
	})

	t.Run("When del ok", func(t *testing.T) {
		l.Del(key)

		allowed, ok := l.IsAllowed(key)
		require.False(t, allowed)
		require.False(t, ok)
	})

	t.Run("When BG sweeper", func(t *testing.T) {
		l.Set(key)

		<-time.After(3 * time.Second)

		allowed, ok := l.IsAllowed(key)
		require.False(t, allowed)
		require.False(t, ok)
	})
}
