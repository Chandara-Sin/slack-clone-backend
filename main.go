package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"slack-clone-api/config"
	"slack-clone-api/domain/user"
	"slack-clone-api/store"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	config.InitConfig()
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.Status(200)
	})

	db := store.CreateDB()
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Println("auto migrate db: ", err)
	}

	u := r.Group("/api")
	u.POST("/users", user.CreateUserHanlder(user.Create(db)))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + viper.GetString("app.port"),
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
}
