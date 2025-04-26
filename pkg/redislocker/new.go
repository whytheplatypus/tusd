package redislocker

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slog"
	"os"
	"time"
)

type LockerOption func(l *RedisLocker)

func WithLogger(logger *slog.Logger) LockerOption {
	return func(l *RedisLocker) {
		l.Logger = logger
	}
}

func New(uri string, lockerOptions ...LockerOption) (*RedisLocker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        uri,
		DialTimeout: 20 * time.Second,
	})
	fmt.Println("redis ping:", client.Ping(context.Background()))
	if res := client.Ping(context.Background()); res.Err() != nil {
		return nil, res.Err()
	}
	rs := redsync.New(goredis.NewPool(client))

	locker := &RedisLocker{
		CreateMutex: func(id string) MutexLock {
			return rs.NewMutex(id, redsync.WithExpiry(LockExpiry))
		},
		Exchange: &RedisLockExchange{
			client: client,
		},
	}
	for _, option := range lockerOptions {
		option(locker)
	}
	//defaults
	if locker.Logger == nil {
		h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
		slog.SetDefault(slog.New(h))
		locker.Logger = slog.Default()
	}

	return locker, nil
}
