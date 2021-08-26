package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExecWebHtml(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}
