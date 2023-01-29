package wizard

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
)

const commandStatePrefix = "command.state.user."

type RedisStateStorage struct {
	RDB *redis.Client
}

type StateStorage interface {
	GetCurrentState(uid int64, dest Wizard) error
	SaveState(uid int64, wizard Wizard) error
	Close() error
}

var commandStateTTL time.Duration
var ctx = context.Background()

func init() {
	var err error
	commandStateTTL, err = time.ParseDuration(os.Getenv("COMMAND_STATE_TTL"))
	if err != nil {
		log.Errorln(err)
	}
}

func ConnectToRedis(options *redis.Options) RedisStateStorage {
	rdb := redis.NewClient(options)
	status := rdb.Ping(ctx)
	if status.Err() != nil {
		panic(status.Err())
	}
	return RedisStateStorage{RDB: rdb}
}

func (rss RedisStateStorage) GetCurrentState(uid int64, dest Wizard) error {
	cmd := rss.RDB.Get(ctx, commandStatePrefix+strconv.FormatInt(uid, 10))
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
	status := rss.RDB.Set(ctx, key, jsonPayload, commandStateTTL)
	return status.Err()
}

func (rss RedisStateStorage) Close() error {
	return rss.RDB.Close()
}
