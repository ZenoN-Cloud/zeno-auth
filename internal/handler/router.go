package handler

import (
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authService *service.AuthService,
	userService *service.UserService,
	jwtManager *token.JWTManager,
) *gin.Engine {
	r := gin.New()
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	authHandler := NewAuthHandler(authService)
	userHandler := NewUserHandler(userService)
	jwksHandler := NewJWKSHandler(jwtManager)

	r.GET("/health", Health)
	r.GET("/jwks", jwksHandler.GetJWKS)

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
		auth.POST("/logout", AuthMiddleware(jwtManager), authHandler.Logout)
	}

	r.GET("/me", AuthMiddleware(jwtManager), userHandler.GetProfile)

	return r
}