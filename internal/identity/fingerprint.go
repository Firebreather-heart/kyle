package identity

import (
	"context"
	"time"
	"github.com/redis/go-redis/v9"
)

type User struct {
	Fingerprint string `json:"fingerprint"`
	ShadowCookie string `json:"-"`
	DocsGeneratedToday int `json:"docs_generated_today"`
	DailyLimitReached bool `json:"daily_limit_reached"`
}

/*
	Triangulate checks for the user in the store using the fingerprint, and cookie.
	It returns the user if found, otherwise it returns nil.
*/
type Service interface{
	Triangulate(ctx context.Context, fingerprint string, cookie string)(*User, error)
	AllowRequest(ctx context.Context, fingerprint string, cookie string) (bool, error)
	AllowLLMCall(ctx context.Context, fingerprint string, cookie string) (bool, error)
	RecordLLMUsage(ctx context.Context, fingerprint string, cookie string) error
	CreateOrUpdateUser(ctx context.Context, user User, ttl time.Duration) error
	PublishTaskUpdate(ctx context.Context, taskID string, status string) error
	SubscribeTask(ctx context.Context, taskID string) *redis.PubSub
	SetTaskFile(ctx context.Context, taskID string, filePath string) error
	GetTaskFile(ctx context.Context, taskID string) (string, error)
	SetSystemStatus(ctx context.Context, provider string, status string, ttl time.Duration) error
	GetSystemStatus(ctx context.Context) (map[string]string, error)
	GetActiveKey(ctx context.Context, provider string, keys []string) (string, error)
	RotateKey(ctx context.Context, provider string, keys []string) (string, error)
}

func NewUser(fingerprint string, cookie string) (*User, error){
	return &User{
		Fingerprint: fingerprint,
		ShadowCookie: cookie,
		DocsGeneratedToday: 0,
		DailyLimitReached: false,
	}, nil
}