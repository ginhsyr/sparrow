package router

import (
	"Sparrow/internal/handler"
	"Sparrow/internal/middleware"
	"Sparrow/internal/model"
	"Sparrow/internal/repository"
	"Sparrow/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"

	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	postRepo := repository.NewPostRepository(db)
	postService := service.NewPostService(postRepo)
	postHandler := handler.NewPostHandler(postService)

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
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
                postLikeGroup := posts.Group("/like")
                {
                    postLikeGroup.POST("", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.PostLike)
                    postLikeGroup.POST("/", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), postHandler.PostLike)
                }
            }

        subscribeGroup := v1.Group("/subscribe")
        {
            subscribeGroup.GET("/notify", middleware.JWTAuth(), middleware.RequireRole(model.Users, model.Admin), handler.Notifies)
        }
		}

	}

	return r
}
