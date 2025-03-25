package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/jwt"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/storage"
	"io"
	"net/http"
	"strings"
	"time"
)

type WithdrawReq struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

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

func WithdrawBalance(c *gin.Context, cfg *config.Config) {
	if !strings.HasPrefix(c.Request.Header.Get("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, "Content-type must be application/json")
		return
	}

	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "you are not auth")
		return
	}
	claims := user.(*jwt.Claims)
	ctx := c.Request.Context()
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Body is nil")
		return
	}
	var WithdrawInfo WithdrawReq

	err = json.Unmarshal(body, &WithdrawInfo)

	if err != nil || WithdrawInfo.Sum == 0 || WithdrawInfo.Order == "" {
		c.JSON(http.StatusBadRequest, "JSON is not correct ")
		return
	}

	if !isValidLuhn(WithdrawInfo.Order) {
		c.JSON(http.StatusUnprocessableEntity, "Not correct number of order")
		return
	}

	err = cfg.Store.WithdrawBalance(ctx, claims.UserID, WithdrawInfo.Sum)
	if err != nil {
		if errors.Is(err, storage.ErrBalanceZero) {
			c.JSON(http.StatusPaymentRequired, "Мало мредств на счету")
			return
		}
		c.JSON(http.StatusInternalServerError, "Error")
		return
	}
	err = cfg.Store.CreateWithdraw(ctx, claims.UserID, WithdrawInfo.Order, WithdrawInfo.Sum, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error")
		return
	}
	c.JSON(http.StatusOK, "Списание прошло успешно!")
}

func GetWithdrawals(c *gin.Context, cfg *config.Config) {
	user, exist := c.Get("user")
	if !exist {
		c.JSON(http.StatusUnauthorized, "You are not auth")
		return
	}
	claims := user.(*jwt.Claims)
	ctx := c.Request.Context()
	withdrawls, err := cfg.Store.GetWithdraws(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Error")
	}
	if len(withdrawls) == 0 {
		c.JSON(http.StatusNoContent, "You have not any withdrawls")
		return
	}

	c.JSON(http.StatusOK, withdrawls)
}
