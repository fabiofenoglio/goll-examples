package main

import (
	"fmt"
	"time"

	gingoll "github.com/fabiofenoglio/gin-goll"
	goll "github.com/fabiofenoglio/goll"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	limiter, _ := goll.New(&goll.Config{
		MaxLoad:    100,
		WindowSize: 10 * time.Second,
	})

	// instantiate the limiter middleware using the limiter instance just created.
	// we assign a default load of "1" per route
	ginLimiter := gingoll.NewLimiterMiddleware(gingoll.Config{
		// Limiter is the goll.LoadLimiter instance
		Limiter: limiter,

		// DefaultRouteLoad is the default load per each route
		// when no route-specific configuration is available
		DefaultRouteLoad: 1,

		// TenantKeyFunc extracts the tenant key from the request context.
		//
		// For instance you can return the request origin IP
		// if you want to limit the load on a per-IP basis,
		// or you could return the username/id of an authenticated client.
		//
		// If you have a single tenant or want to limit globally
		// you can return a fixed string or use the TenantKey parameter instead.
		TenantKeyFunc: func(c *gin.Context) (string, error) {
			// we will limit the load on a per-ip basis
			return c.ClientIP(), nil
		},

		// AbortHandler decides how we respond when a request
		// exceeds the load limit
		AbortHandler: func(c *gin.Context, result goll.SubmitResult) {
			if result.RetryInAvailable {
				c.Header("X-Retry-In", fmt.Sprintf("%v", result.RetryIn.Milliseconds()))
				c.String(429, fmt.Sprintf("Too much! retry in %v ms", result.RetryIn.Milliseconds()))
				c.Abort()
			} else {
				c.AbortWithStatus(429)
			}
		},
	})

	// let's plug in the load limiter for a single route with a load of 20
	r.GET("/", ginLimiter.WithLoad(20), func(c *gin.Context) {
		c.String(200, "ok!")
	})

	fmt.Println("now listening. try making six or more rapid requests to http://localhost:9000")

	err := r.Run(":9000")
	if err != nil {
		panic(err)
	}
}
