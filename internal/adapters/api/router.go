package api

import (
	"focus-dev-challenge/internal/core/ports"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	Engine  *gin.Engine
	service ports.AppService
}

func NewRouter(svc ports.AppService, logger *zap.Logger, debug bool) *Router {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, true))

	router := Router{
		Engine:  engine,
		service: svc,
	}

	router.attachRoutes()
	return &router
}

func (r *Router) attachRoutes() {
	v1 := r.Engine.Group("/")
	{
		v1.GET("campaigns", r.GetCampaigns)
		v1.POST("campaigns", r.CreateCampaign)
		v1.GET("campaigns/:id", r.GetCampaign)
		v1.POST("campaigns/:id/send", r.SendCampaign)
		v1.POST("campaigns/:id/personalized-preview", r.Preview)
	}
}
