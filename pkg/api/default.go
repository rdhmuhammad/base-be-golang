package api

import (
	"base-be-golang/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func Default() *Api {
	server := gin.Default()

	server.Use(middleware.AllowCORS())

	var routers = []Router{}

	return &Api{
		server:  server,
		routers: routers,
	}
}
