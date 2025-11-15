package handler

import (
	"net/http"

	"github.com/ZenoN-Cloud/zeno-auth/internal/token"
	"github.com/gin-gonic/gin"
)

type JWKSHandler struct {
	jwtManager *token.JWTManager
}

func NewJWKSHandler(jwtManager *token.JWTManager) *JWKSHandler {
	return &JWKSHandler{jwtManager: jwtManager}
}

func (h *JWKSHandler) GetJWKS(c *gin.Context) {
	jwks, err := h.jwtManager.GetJWKS(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to get JWKS"})
		return
	}

	c.JSON(http.StatusOK, jwks)
}