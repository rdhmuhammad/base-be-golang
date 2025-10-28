package api

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/db"
	"base-be-golang/pkg/middleware"
	"base-be-golang/pkg/miniostorage"
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

func Default() *Api {
	// Initialize Sentry
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv(""),
		Environment:      os.Getenv("ENVIRONMENT"),
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Add source context for better error tracing
			return event
		},
	})
	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	server := gin.Default()

	server.Use(middleware.AllowCORS())

	// Add Sentry middleware with enhanced configuration
	server.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         30 * time.Second,
	}))

	// Add custom Sentry middleware for request enrichment
	server.Use(middleware.SentryMiddleware())

	_, err = db.Default()
	if err != nil {
		panic(fmt.Sprintf("panic at db connection: %s", err.Error()))
	}

	_ = cache.Default()

	_ = miniostorage.NewConnection(miniostorage.Conn{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
	})

	var routers = []Router{}

	middleware.CaptureErrorUsecase(context.Background(), fmt.Errorf("Start application"))

	return &Api{
		server:  server,
		routers: routers,
	}
}
