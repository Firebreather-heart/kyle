package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/firebreather-heart/kyle/internal/identity"
)

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