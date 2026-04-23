package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/firebreather-heart/kyle/internal/identity"
)

type contextKey string
const userKey contextKey = "user"


func Logger(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		start := time.Now()
		log.Printf("%s %s %s %s", start, r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s took %s", start, r.Method, r.URL.Path, time.Since(start))
	})
}

func CORS(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Fingerprint, Cookie")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authenticator(idService identity.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request){
				fp := r.Header.Get("X-Fingerprint")
				var cookieVal string
				if c, err := r.Cookie("kyle_id"); err == nil {
					cookieVal = c.Value
				}
				
				user, err := idService.Triangulate(r.Context(), fp, cookieVal)
				if err != nil {
					log.Printf("Authentication error: %v", err)
				}
				ctx := context.WithValue(r.Context(), userKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
	}
}

func RateLimiter(idService identity.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request){
				fp := r.Header.Get("X-Fingerprint")
				var cookieVal string
				if c, err := r.Cookie("kyle_id"); err == nil {
					cookieVal = c.Value
				}
				
				allowed, err := idService.AllowRequest(r.Context(), fp, cookieVal)
				if err != nil {
					log.Printf("Rate limiting error: %v", err)
				}
				if !allowed {
					http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
					return
				}
				next.ServeHTTP(w, r)
			})
	}
}

func GetUser(r *http.Request) *identity.User {
	if user, ok := r.Context().Value(userKey).(*identity.User); ok {
		return user
	}
	return nil
}