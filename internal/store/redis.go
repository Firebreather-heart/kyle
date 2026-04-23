package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/firebreather-heart/kyle/internal/identity"
)

const MAX_REQUESTS_PER_MINUTE = 10
const MAX_DOCS_PER_DAY = 2

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error){
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err:= client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}
	return &RedisStore{client: client}, nil
}

/*
	Triangulate checks for the user in the redis store using the fingerprint, and cookie.
	It returns the user if found, otherwise it returns nil.
*/
func (r *RedisStore) Triangulate(ctx context.Context, fingerprint string, cookie string) (*identity.User, error) {
	aliasedFP, _ := r.client.Get(ctx, fmt.Sprintf("cookie:%s", cookie)).Result()
	targetFP := fingerprint
	if aliasedFP != "" { targetFP = aliasedFP }

	val, err := r.client.Get(ctx, fmt.Sprintf("fp:%s", targetFP)).Result()
	if err == redis.Nil { return nil, nil }
	if err != nil { return nil, err }

	var user identity.User
	json.Unmarshal([]byte(val), &user)

	// Fetch dynamic usage from Redis-native TTL key
	usageKey := fmt.Sprintf("usage:v1:%s", targetFP)
	usage, _ := r.client.Get(ctx, usageKey).Int()
	user.DocsGeneratedToday = usage
	user.DailyLimitReached = usage >= MAX_DOCS_PER_DAY

	return &user, nil
}

func (r *RedisStore) CreateOrUpdateUser(ctx context.Context, user identity.User, ttl time.Duration) error {
	userdata, _ := json.Marshal(user)
	fpKey := fmt.Sprintf("fp:%s", user.Fingerprint)
	cookieKey := fmt.Sprintf("cookie:%s", user.ShadowCookie)

	pipe := r.client.Pipeline()
	pipe.Set(ctx, fpKey, userdata, ttl)
	pipe.Set(ctx, cookieKey, user.Fingerprint, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisStore) getOrCreateUser(ctx context.Context, fingerprint, cookie string) (*identity.User, error) {
	user, _ := r.Triangulate(ctx, fingerprint, cookie)
	if user == nil {
		user, _ = identity.NewUser(fingerprint, cookie)
		r.CreateOrUpdateUser(ctx, *user, 30*24*time.Hour)
	}
	return user, nil
}

func (r *RedisStore) AllowLLMCall(ctx context.Context, fingerprint string, cookie string) (bool, error) {
	user, _ := r.getOrCreateUser(ctx, fingerprint, cookie)
	usageKey := fmt.Sprintf("usage:v1:%s", user.Fingerprint)
	usage, _ := r.client.Get(ctx, usageKey).Int()
	return usage < MAX_DOCS_PER_DAY, nil
}

func (r *RedisStore) RecordLLMUsage(ctx context.Context, fp string, cookie string) error {
	user, _ := r.getOrCreateUser(ctx, fp, cookie)
	usageKey := fmt.Sprintf("usage:v1:%s", user.Fingerprint)
	
	count, err := r.client.Incr(ctx, usageKey).Result()
	if err == nil && count == 1 {
		r.client.Expire(ctx, usageKey, 24*time.Hour)
	}
	return err
}

func (r *RedisStore) AllowRequest(ctx context.Context, fingerprint string, cookie string) (bool, error) {
	user, _ := r.getOrCreateUser(ctx, fingerprint, cookie)
	limitKey := fmt.Sprintf("limit:%s", user.Fingerprint)
	count, _ := r.client.Incr(ctx, limitKey).Result()
	if count == 1 { r.client.Expire(ctx, limitKey, time.Minute) }
	return count <= MAX_REQUESTS_PER_MINUTE, nil
}

func (r *RedisStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil { return "", nil }
	return val, err
}

func (r *RedisStore) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) SetTaskFile(ctx context.Context, taskID string, filePath string) error {
	return r.client.Set(ctx, fmt.Sprintf("task:file:%s", taskID), filePath, 24*time.Hour).Err()
}

func (r *RedisStore) GetTaskFile(ctx context.Context, taskID string) (string, error) {
	val, err := r.client.Get(ctx, fmt.Sprintf("task:file:%s", taskID)).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisStore) SetSystemStatus(ctx context.Context, provider string, status string, ttl time.Duration) error {
	return r.client.Set(ctx, fmt.Sprintf("status:%s", provider), status, ttl).Err()
}

func (r *RedisStore) GetSystemStatus(ctx context.Context) (map[string]string, error) {
	keys, err := r.client.Keys(ctx, "status:*").Result()
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, key := range keys {
		val, _ := r.client.Get(ctx, key).Result()
		provider := key[7:] // remove "status:"
		res[provider] = val
	}
	return res, nil
}

func (r *RedisStore) PublishTaskUpdate(ctx context.Context, taskID string, status string) error {
	channel := fmt.Sprintf("task:%s", taskID)
	return r.client.Publish(ctx, channel, status).Err()
}

func (r *RedisStore) SubscribeTask(ctx context.Context, taskID string) *redis.PubSub {
	channel := fmt.Sprintf("task:%s", taskID)
	return r.client.Subscribe(ctx, channel)
}

func (r *RedisStore) GetActiveKey(ctx context.Context, provider string, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("no keys provided for %s", provider)
	}
	key := fmt.Sprintf("idx:%s", provider)
	idx, err := r.client.Get(ctx, key).Int64()
	if err != nil && err != redis.Nil {
		return "", err
	}
	if int(idx) >= len(keys) {
		idx = 0
	}
	return keys[idx], nil
}

func (r *RedisStore) RotateKey(ctx context.Context, provider string, keys []string) (string, error) {
	if len(keys) == 0 {
		return "", fmt.Errorf("no keys provided for %s", provider)
	}
	key := fmt.Sprintf("idx:%s", provider)
	idx, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return "", err
	}
	newIdx := int(idx) % len(keys)
	if int(idx) != newIdx {
		r.client.Set(ctx, key, newIdx, 0)
	}
	return keys[newIdx], nil
}