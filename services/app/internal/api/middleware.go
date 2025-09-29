package api

import (
	"context"
	"hafiztri123/app-link-shortener/internal/auth"
	"hafiztri123/app-link-shortener/internal/response"
	"hafiztri123/app-link-shortener/internal/shared"
	"hpj/hv1-link-shortener/shared/models"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mileusna/useragent"
	"github.com/oschwald/maxminddb-golang"
	"golang.org/x/time/rate"
)

func RateLimiter(r rate.Limit, b int) func(http.Handler) http.Handler {
	limiter := rate.NewLimiter(r, b)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				response.Error(w, http.StatusTooManyRequests, "The API is at capacity, please try again later")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RedisRateLimiter(redisClient *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now().UnixNano()
			windowStart := now - window.Nanoseconds()

			pipe := redisClient.TxPipeline()
			pipe.ZRemRangeByScore(r.Context(), ip, "0", strconv.FormatInt(windowStart, 10))
			pipe.ZAdd(r.Context(), ip, &redis.Z{Score: float64(now), Member: now})

			pipe.ZCard(r.Context(), ip)
			pipe.Expire(r.Context(), ip, window)

			cmds, err := pipe.Exec(r.Context())

			if err != nil {
				slog.Warn("Rate limiter failed", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			countCmd := cmds[2].(*redis.IntCmd)

			if countCmd.Val() > int64(limit) {
				response.Error(w, http.StatusTooManyRequests, "Too many requests")
				return
			}

			next.ServeHTTP(w, r)

		})
	}

}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

func AuthMiddleware(ts *auth.TokenService, permissive bool) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if permissive && authHeader == "" {
				h.ServeHTTP(w, r)
				return
			}

			if authHeader == "" {
				slog.Error("missing authorization header")
				response.Error(w, http.StatusUnauthorized, "authorization header required")
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			//Authorization must have the prefix "Bearer"
			if tokenString == authHeader {
				slog.Error("missing 'Bearer' in auth header")
				response.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			claims, err := ts.ValidateToken(tokenString)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), shared.UserContextKey, claims)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func MetadataMiddleware(db *maxminddb.Reader) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var ipStr string

			if fwdAdress := r.Header.Get("X-Forwarded-For"); fwdAdress != "" {
				ipStr = strings.TrimSpace(strings.Split(fwdAdress, ",")[0])
			} else {
				if tempIp, _, err := net.SplitHostPort(r.RemoteAddr); err != nil {
					ipStr = r.RemoteAddr
				} else {
					ipStr = tempIp
				}
			}

			ip := net.ParseIP(ipStr)

			ua := useragent.Parse(r.Header.Get("User-Agent"))
			var deviceType string

			if ua.Mobile {
				deviceType = "Mobile"
			} else if ua.Desktop {
				deviceType = "Desktop"
			} else if ua.Tablet {
				deviceType = "Tablet"
			} else {
				deviceType = "Unknown"
			}

			var geoData models.GeoIPCity
			country, city := "unknown", "unknown"

			if err := db.Lookup(ip, &geoData); err == nil {
				country = geoData.Country.ISOCode
				if cityName, ok := geoData.City.Names["en"]; ok {
					city = cityName
				}
			}

			clickData := &models.Click{
				Timestamp: time.Now().UTC(),
				Path:      r.URL.Path,
				IPAddress: ipStr,
				Referer:   r.Header.Get("Referer"),
				UserAgent: r.Header.Get("User-Agent"),
				Device:    deviceType,
				OS:        ua.OS,
				Browser:   ua.Name,
				Country:   country,
				City:      city,
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), shared.ClickDataKey, clickData)))
		})
	}
}
