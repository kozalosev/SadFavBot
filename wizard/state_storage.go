package wizard

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const commandStatePrefix = "command.state.user."

type RedisStateStorage struct {
	rdb *redis.Client
	ttl time.Duration
	ctx context.Context
}

type StateStorage interface {
	GetCurrentState(uid int64, dest Wizard) error
	SaveState(uid int64, wizard Wizard) error
	Close() error
}

func ConnectToRedis(ctx context.Context, ttl time.Duration, options *redis.Options) RedisStateStorage {
	rdb := redis.NewClient(options)
	status := rdb.Ping(ctx)
	if status.Err() != nil {
		panic(status.Err())
	}
	return RedisStateStorage{
		rdb: rdb,
		ttl: ttl,
		ctx: ctx,
	}
}

func (rss RedisStateStorage) GetCurrentState(uid int64, dest Wizard) error {
	cmd := rss.rdb.Get(rss.ctx, commandStatePrefix+strconv.FormatInt(uid, 10))
	if cmd.Err() != nil {
		return cmd.Err()
	}
	if err := json.Unmarshal([]byte(cmd.Val()), dest); err != nil {
		return err
	}
	return nil
}

func (rss RedisStateStorage) SaveState(uid int64, wizard Wizard) error {
	payload, err := json.Marshal(wizard)
	if err != nil {
		return err
	}

	jsonPayload := string(payload)
	key := commandStatePrefix + strconv.FormatInt(uid, 10)
	status := rss.rdb.Set(rss.ctx, key, jsonPayload, rss.ttl)
	return status.Err()
}

func (rss RedisStateStorage) Close() error {
	return rss.rdb.Close()
}
