// Package api -----------------------------
// @file      : health.go
// @author    : 被测试耽误的大厨
// @contact   : 13383088061@163.com
// -------------------------------------------
package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
