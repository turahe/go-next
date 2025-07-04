package middleware

import (
	"github.com/gin-gonic/gin"
	limiterpkg "github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimitMiddleware() gin.HandlerFunc {
	rate, _ := limiterpkg.NewRateFromFormatted("100-M") // 100 requests per minute
	store := memory.NewStore()
	middleware := ginlimiter.NewMiddleware(limiterpkg.New(store, rate))
	return middleware
}
