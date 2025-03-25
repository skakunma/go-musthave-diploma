package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
	"net/http"
)

func GetBalance(c *gin.Context, cfg *config.Config) {
	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not auth")
		return
	}
	claims := user.(*jwt.Claims)
	ctx := c.Request.Context()
	balance, err := cfg.Store.GetBalance(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "IDK")
		return
	}
	c.JSON(http.StatusOK, balance)

}
