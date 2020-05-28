package api

import (
	"context"
	"log"
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
	r := gin.Default()
	s := &ApiServer{
		ctxt: ctxt,
		conf: conf,
		eng:  r,
	}

	r.GET("/config", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"data": s.conf,
		})
	})
	r.PUT("/config", func(c *gin.Context) {
		c.JSON(501, gin.H{})
	})

	r.GET("/routes", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	return s
}

func (s *ApiServer) Start() {
	go func() {
		server := endless.NewServer(s.conf.APIAddress, s.eng)

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
