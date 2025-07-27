package main

import (
	"compus_blog/basic/internal/pkg/saramax"
	"github.com/gin-gonic/gin"
)

type App struct {
	server    *gin.Engine
	consumers []saramax.Consumer
}
