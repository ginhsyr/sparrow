package main

import (
	"Sparrow/configs"
	"Sparrow/internal/router"
	"Sparrow/internal/utils"
	"Sparrow/internal/utils/database"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	if err := utils.InitLogger(); err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer utils.Log.Sync()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai", configs.DBConfig.Host, configs.DBConfig.User, configs.DBConfig.Password, configs.DBConfig.DBName, configs.DBConfig.Port)
	utils.Log.Info(
		"Database config loaded",
		zap.String("host", configs.DBConfig.Host),
		zap.String("dbName", configs.DBConfig.DBName),
		zap.String("port", configs.DBConfig.Port),
	)
	var db *gorm.DB

	var err error
	for i := 0; i < 20; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true})
		if err == nil {
			break
		}
		utils.Log.Info("Database connection failed, retrying", zap.Int("times", i+1))
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		utils.Log.Fatal("Could not connect to database", zap.Error(err))
	}
	database.DBInit(db)

	r := router.SetupRouter(db)
	addr := configs.ServerConfig.Port
	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", addr),
		Handler:           r,
		ReadTimeout:       configs.ServerConfig.ReadTimeout,
		WriteTimeout:      configs.ServerConfig.WriteTimeout,
		IdleTimeout:       configs.ServerConfig.IdleTimeout,
		ReadHeaderTimeout: configs.ServerConfig.ReadHeaderTimeout,
	}

	utils.Log.Info(
		"HTTP server starting",
		zap.String("addr", server.Addr),
		zap.Duration("readTimeout", configs.ServerConfig.ReadTimeout),
		zap.Duration("writeTimeout", configs.ServerConfig.WriteTimeout),
		zap.Duration("idleTimeout", configs.ServerConfig.IdleTimeout),
		zap.Duration("shutdownTimeout", configs.ServerConfig.ShutdownTimeout),
	)
	fmt.Printf("Server running at http://localhost:%s\n", addr)

	errCh := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		utils.Log.Info("Shutdown signal received", zap.String("signal", sig.String()))
	case err := <-errCh:
		utils.Log.Fatal("failed to run server", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), configs.ServerConfig.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		utils.Log.Error("Graceful shutdown failed", zap.Error(err))
		if closeErr := server.Close(); closeErr != nil {
			utils.Log.Error("Forced server close failed", zap.Error(closeErr))
		}
	} else {
		utils.Log.Info("Server shutdown complete")
	}
}
