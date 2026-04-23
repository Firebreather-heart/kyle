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

func NewRedisStore(url string) (*RedisStore, error){
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
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
	aliasedFP, err := r.client.Get(ctx, fmt.Sprintf("cookie:%s", cookie)).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check cookie alias: %w", err)
	}

	targetFP := fingerprint
	if aliasedFP != "" {
		targetFP = aliasedFP
	}

	val, err := r.client.Get(ctx, fmt.Sprintf("fp:%s", targetFP)).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to fetch user account: %w", err)
	}

	var user identity.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}
	return &user, nil
}


func (r *RedisStore) CreateOrUpdateUser(ctx context.Context, user identity.User, ttl time.Duration) error {
	userdata, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	fpKey := fmt.Sprintf("fp:%s", user.Fingerprint)
	cookieKey := fmt.Sprintf("cookie:%s", user.ShadowCookie)

	// Attempt to find existing user first to check for conflicts
	existingUser, err := r.Triangulate(ctx, user.Fingerprint, user.ShadowCookie)
	if err != nil {
		return fmt.Errorf("failed to triangulate user: %w", err)
	}

	pipe := r.client.Pipeline()

	if existingUser == nil {
		pipe.Set(ctx, fpKey, userdata, ttl)
		pipe.Set(ctx, cookieKey, user.Fingerprint, ttl)
	} else {
		pipe.Set(ctx, fpKey, userdata, ttl)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

func (r *RedisStore) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStore) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("failed to get key: %w", err)
	}
	return val, nil
}

func (r *RedisStore) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStore) getOrCreateUser(ctx context.Context, fingerprint, cookie string) (*identity.User, error) {
	user, err := r.Triangulate(ctx, fingerprint, cookie)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user: %w", err)
	}

	if user == nil {
		user, _ = identity.NewUser(fingerprint, cookie)
		if err := r.CreateOrUpdateUser(ctx, *user, 24*time.Hour); err != nil {
			return nil, fmt.Errorf("failed to create new user: %w", err)
		}
	}
	return user, nil
}

// AllowRequest checks the traffic limit (requests per minute)
func (r *RedisStore) AllowRequest(ctx context.Context, fingerprint string, cookie string) (bool, error) {
	user, err := r.getOrCreateUser(ctx, fingerprint, cookie)
	if err != nil {
		return false, err
	}

	limitKey := fmt.Sprintf("limit:%s", user.Fingerprint)
	count, err := r.client.Incr(ctx, limitKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to increment request count: %w", err)
	}

	if count == 1 {
		r.client.Expire(ctx, limitKey, time.Minute)
	}

	return count <= MAX_REQUESTS_PER_MINUTE, nil
}

// AllowLLMCall checks the business quota (docs per day)
func (r *RedisStore) AllowLLMCall(ctx context.Context, fingerprint string, cookie string) (bool, error) {
	user, err := r.getOrCreateUser(ctx, fingerprint, cookie)
	if err != nil {
		return false, err
	}

	// Check the daily limits stored in the user struct
	if user.DailyLimitReached || user.DocsGeneratedToday >= MAX_DOCS_PER_DAY {
		return false, nil
	}

	return true, nil
}

func (r RedisStore) RecordLLMUsage(ctx context.Context, fp string, cookie string) error {
	user, err := r.getOrCreateUser(ctx, fp, cookie)
	if err != nil {
		return fmt.Errorf("failed to get or create user: %w", err)
	}
	user.DocsGeneratedToday++
	if user.DocsGeneratedToday >= MAX_DOCS_PER_DAY {
		user.DailyLimitReached = true
	}
	if err := r.CreateOrUpdateUser(ctx, *user, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *RedisStore) PublishTaskUpdate(ctx context.Context, taskID string, status string) error {
	channel := fmt.Sprintf("task:%s", taskID)
	return r.client.Publish(ctx, channel, status).Err()
}

func (r *RedisStore) SubscribeTask(ctx context.Context, taskID string) *redis.PubSub {
	channel := fmt.Sprintf("task:%s", taskID)
	return r.client.Subscribe(ctx, channel)
}