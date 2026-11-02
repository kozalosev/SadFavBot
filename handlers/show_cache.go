package handlers

import (
	"context"
	"encoding/json"
	"errors"
	log "log/slog"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kozalosev/SadFavBot/db/dto"
	"github.com/kozalosev/goSadTgBot/logconst"
)

const showCacheKeyPrefix = "show:"

var (
	// ErrShowCacheMiss is returned by ShowCache.Get if the entry doesn't exist or has expired.
	ErrShowCacheMiss = errors.New("show cache miss")

	showCacheTTL time.Duration
)

func init() {
	ttlSecs := 3600 // default
	if env := os.Getenv("SHOW_CACHE_TTL_SECS"); len(env) > 0 {
		if val, err := strconv.Atoi(env); err != nil {
			log.Error("Failed to parse an environment variable",
				logconst.FieldFunc, "init",
				logconst.FieldConst, "SHOW_CACHE_TTL_SECS",
				logconst.FieldError, err)
		} else {
			ttlSecs = val
		}
	}
	showCacheTTL = time.Duration(ttlSecs) * time.Second
}

// ShowEntry is the list of favs found by the /show command, bound to the user who has requested it.
type ShowEntry struct {
	OwnerUID int64
	Favs     []*dto.Fav
}

// ShowCache is a short-living storage of ShowEntry items, which lets the buttons of /show work without the database.
type ShowCache interface {
	Put(ctx context.Context, token string, entry *ShowEntry) error
	Get(ctx context.Context, token string) (*ShowEntry, error)
}

// RedisShowCache is the Redis implementation of ShowCache. Reading doesn't prolong the lifetime of an entry.
// The struct holds a single reference, so it's passed by value.
type RedisShowCache struct {
	rdb *redis.Client
}

// NewRedisShowCache connects to Redis and panics if the server is unavailable, like the state storage does.
func NewRedisShowCache(ctx context.Context, options *redis.Options) RedisShowCache {
	rdb := redis.NewClient(options)
	if err := rdb.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	return RedisShowCache{rdb: rdb}
}

func (cache RedisShowCache) Close() error {
	return cache.rdb.Close()
}

func (cache RedisShowCache) Put(ctx context.Context, token string, entry *ShowEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return cache.rdb.Set(ctx, showCacheKeyPrefix+token, data, showCacheTTL).Err()
}

func (cache RedisShowCache) Get(ctx context.Context, token string) (*ShowEntry, error) {
	data, err := cache.rdb.Get(ctx, showCacheKeyPrefix+token).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrShowCacheMiss
	}
	if err != nil {
		return nil, err
	}
	entry := new(ShowEntry)
	if err = json.Unmarshal(data, entry); err != nil {
		return nil, err
	}
	return entry, nil
}
