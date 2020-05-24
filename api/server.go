package api

import (
	"context"
	"log"
	"path"
	"runtime"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	"github.com/thomas-neuman/spigot/config"
)

type ApiServer struct {
	ctxt context.Context
	conf *config.Configuration
	eng  *gin.Engine
}

func NewApiServer(ctxt context.Context, conf *config.Configuration) *ApiServer {
	s := &ApiServer{
		ctxt: ctxt,
		conf: conf,
	}

	r := gin.Default()
	r.GET("/config", func(c *gin.Context) {
		_, err := conf.Get("base")
		if err != nil {
			c.JSON(500, gin.H{})
		}
		c.JSON(200, gin.H{})
	})

	return s
}

func (s *ApiServer) Start() {
	go func() {
		server := endless.NewServer(":8788", s.eng)

		go func() {
			err := server.ListenAndServe()
			if err != nil {
				return
			}
		}()

		<-s.ctxt.Done()
		log.Println("Shutting down API server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			log.Println("Forcing API server shutdown...")
		}
		log.Println("API server shut down")
	}()
}
