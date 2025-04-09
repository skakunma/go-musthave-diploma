package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	jwtauth "github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
)

type (
	AuthForm struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
)

func HashPassword(cfg *config.Config, password string) string {
	hash := sha256.Sum256([]byte(password + cfg.Salt))
	return hex.EncodeToString(hash[:])
}

func RegisterHandler(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content must be application/json"})
		return
	}

	var infoUser AuthForm

	if err := c.ShouldBindJSON(&infoUser); err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	if infoUser.Login == "" || infoUser.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login and password are required"})
		return
	}

	hashPassword := HashPassword(cfg, infoUser.Password)
	password := infoUser.Password
	infoUser.Password = hashPassword

	ctx := c.Request.Context()
	exist, err := cfg.Store.IsUserExist(ctx, infoUser.Login)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Database error"})
		return
	}
	if exist {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	err = cfg.Store.CreateUser(ctx, infoUser.Login, infoUser.Password)
	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Could not create user"})
		return
	}

	id, err := cfg.Store.GetID(ctx, infoUser.Login)

	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Problem with service")
		return
	}

	token, err := jwtauth.BuildJWTString(id)

	if err != nil {
		cfg.Sugar.Error(err)
		c.JSON(http.StatusBadGateway, "Problem With Service")
		return
	}

	c.SetCookie("jwt", token, 3600, "/", "", false, false)
	jwtHeader := fmt.Sprintf("Bearer %v", token)

	c.Writer.Header().Set("Authorization", jwtHeader)

	c.JSON(http.StatusOK, gin.H{"login": infoUser.Login, "password": password})
}
