package app

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/Halturshik/TicketAgregator-API/internal/auth"
	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/common/httpx"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/ratelimit"
)

type RateLimitMiddleware struct {
	limiter *ratelimit.Limiter
}

func NewRateLimitMiddleware(limiter *ratelimit.Limiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: limiter}
}

func (m *RateLimitMiddleware) Search(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		policy := guestSearchPolicy
		identity := "ip:" + clientIP(r)
		if userID, ok := auth.UserIDFromContext(r.Context()); ok {
			policy = userSearchPolicy
			identity = "user:" + strconv.Itoa(userID)
		}
		m.handle(policy, identity, next, w, r)
	})
}

func (m *RateLimitMiddleware) SearchPage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handle(searchPagePolicy, requestIdentity(r), next, w, r)
	})
}

func (m *RateLimitMiddleware) PublicLookup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handle(publicLookupPolicy, "ip:"+clientIP(r), next, w, r)
	})
}

func (m *RateLimitMiddleware) handle(
	policy ratelimit.Policy,
	identity string,
	next http.Handler,
	w http.ResponseWriter,
	r *http.Request,
) {
	decision, err := m.limiter.Allow(r.Context(), policy, identity)
	if err != nil {
		logger.Error("Ошибка проверки ограничения запросов policy=%s: %v", policy.Name, err)
		_ = httpx.WriteJSON(w, apierror.ErrInternal.Status, apierror.ErrInternal)
		return
	}
	if !decision.Allowed {
		retryAfter := int((decision.RetryAfter + time.Second - 1) / time.Second)
		if retryAfter < 1 {
			retryAfter = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		logger.Warn("Превышено ограничение запросов policy=%s identity=%s", policy.Name, identity)
		_ = httpx.WriteJSON(w, apierror.ErrRateLimited.Status, apierror.ErrRateLimited)
		return
	}
	next.ServeHTTP(w, r)
}

func requestIdentity(r *http.Request) string {
	if userID, ok := auth.UserIDFromContext(r.Context()); ok {
		return "user:" + strconv.Itoa(userID)
	}
	return "ip:" + clientIP(r)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}
