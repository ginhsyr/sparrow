package main

import (
	"Sparrow/configs"
	"Sparrow/internal/router"
	"Sparrow/internal/utils"
	"Sparrow/internal/utils/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func main() {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	_ = utils.InitLogger()
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
	fmt.Printf("Server running at http://localhost:%s\n", addr)
	if err := r.Run(fmt.Sprintf(":%s", addr)); err != nil {
		utils.Log.Fatal("failed to run server", zap.Error(err))
	}
}
