package handler

import (
	"github.com/ZenoN-Cloud/zeno-auth/internal/service"
	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authService service.AuthServiceInterface,
	userService service.UserServiceInterface,
	jwtManager *token.JWTManager,
) *gin.Engine {
	r := gin.New()
	r.Use(LoggingMiddleware())
	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())

	// Always add health endpoint
	r.GET("/health", Health)
	r.HEAD("/health", Health)
	r.GET("/debug", Debug)

	// Only add other endpoints if services are available
	if authService != nil && userService != nil && jwtManager != nil {
		authHandler := NewAuthHandler(authService)
		userHandler := NewUserHandler(userService)
		jwksHandler := NewJWKSHandler(jwtManager)

		r.GET("/jwks", jwksHandler.GetJWKS)

		auth := r.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", AuthMiddleware(jwtManager), authHandler.Logout)
		}

		r.GET("/me", AuthMiddleware(jwtManager), userHandler.GetProfile)
	}

	return r
}
