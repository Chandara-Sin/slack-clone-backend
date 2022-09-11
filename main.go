package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"slack-clone-api/auth"
	"slack-clone-api/config"
	"slack-clone-api/domain/user"
	"slack-clone-api/mw"
	"slack-clone-api/store"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	config.InitConfig()
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:8000",
		"http://localhost:3000",
	}
	config.AllowHeaders = []string{
		"Authorization",
	}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "Ok v1")
	})

	db := store.CreateDB()
	if err := db.AutoMigrate(&user.User{}); err != nil {
		log.Println("auto migrate db: ", err)
	}

	r.POST("/api/oauth/token", auth.JWTConfigHandler(user.GetUserByEmail(db), user.GetUser(db)))
	u := r.Group("/api")
	u.Use(mw.JWTConfig(viper.GetString("jwt.secret")))
	u.POST("/users", user.CreateUserHanlder(user.Create(db)))
	u.GET("/users", user.GetUserHanlder(user.GetUser(db)))

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
