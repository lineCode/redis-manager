package server

import (
	"github.com/gin-gonic/gin"
)

func (srv *Server) runTTYServer(router *gin.RouterGroup) {
	router.Any("/", func(context *gin.Context) {
		context.Writer.WriteString("Windows not support tty")
	})
}
