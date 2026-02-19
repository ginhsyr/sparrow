package router

import (
	"Sparrow/configs"
	"Sparrow/internal/handler"
	"Sparrow/internal/middleware"
	"Sparrow/internal/model"
	"Sparrow/internal/repository"
	"Sparrow/internal/service"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"

	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()

	limiter := middleware.NewIPRateLimiter(configs.ServerConfig.RateLimitPerMin, configs.ServerConfig.RateLimitBurst)
	r.Use(middleware.TraceID())
	r.Use(middleware.RequestLogger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS(configs.ServerConfig.CORSAllowOrigins))
	r.Use(limiter.Middleware())

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postService)

	liveHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}

	readyHandler := func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"error":  err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}

	r.GET("/health/live", liveHandler)
	r.GET("/health/ready", readyHandler)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.GET("/health/live", liveHandler)
			v1.GET("/health/ready", readyHandler)

			v1.GET("/ping", func(context *gin.Context) {
				context.JSON(http.StatusOK, "pong")
			})

			userGroup := v1.Group("/user")
			{
				userGroup.GET("/:id", userHandler.GetUser)
			}

			authGroup := v1.Group("/auth")
			{
				authGroup.POST("/register", userHandler.UserRegister)
				authGroup.POST("/login", userHandler.LogIn)
			}

			posts := v1.Group("/posts")
			{
				posts.GET("/:id", postHandler.GetPost)
				// Support both "/posts" and "/posts/" to avoid redirect issues on POST
				posts.POST("", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.CreatePost)
				posts.POST("/", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.CreatePost)
				posts.POST("/:id/like", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.PostLike)
				posts.POST("/:id/like/", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.PostLike)
			}

			subscribeGroup := v1.Group("/subscribe")
			{
				subscribeGroup.GET("/notify", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), handler.Notifies)
			}
		}

	}

	return r
}
