package wizard

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

const (
	commandStatePrefix = "command.state.user."
	noActiveWizardTr   = "wizard.active.not.set"
)

type RedisStateStorage struct {
	rdb *redis.Client
	ttl time.Duration
	ctx context.Context
}

type StateStorage interface {
	GetCurrentState(uid int64, dest Wizard) error
	SaveState(uid int64, wizard Wizard) error
	DeleteState(uid int64) error
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
	cmd := rss.rdb.Get(rss.ctx, getRedisStateKey(uid))
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
	status := rss.rdb.Set(rss.ctx, getRedisStateKey(uid), jsonPayload, rss.ttl)
	return status.Err()
}

func (rss RedisStateStorage) DeleteState(uid int64) error {
	cmd := rss.rdb.Del(rss.ctx, getRedisStateKey(uid))
	if cmd.Err() != nil {
		return cmd.Err()
	} else if cmd.Val() == 0 {
		return errors.New(noActiveWizardTr)
	} else {
		return nil
	}
}

func (rss RedisStateStorage) Close() error {
	return rss.rdb.Close()
}

func getRedisStateKey(uid int64) string {
	return commandStatePrefix + strconv.FormatInt(uid, 10)
}
