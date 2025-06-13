package app_server

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"todo-app/internal/todos"
)

var (
	logger *zap.SugaredLogger
	e      *echo.Echo
	db     *gorm.DB
)

func InitializeServer() {
	baseLogger, _ := zap.NewProduction()
	defer baseLogger.Sync() // flushes buffer, if any
	logger = baseLogger.Sugar()

	e = echo.New()
	e.HideBanner = true

	err := initializeDb()
	if err != nil {
		logger.Errorw("failed to initialize database", "error", err)

		os.Exit(1)
	}
	logger.Info("initialized database")

	v := validator.New()

	todoService := todos.GetService(logger, db, v)

	todoEndpointHandler := todos.GetEndpointHandler(
		logger,
		todoService,
		e,
		v,
	)

	todoEndpointHandler.AddEndpoints()

	logger.Fatal(e.Start(":8765"))
}

func initializeDb() error {
	dbString := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", "root", "root", "mysql", "todo")
	var err error
	db, err = gorm.Open(mysql.Open(dbString), &gorm.Config{})
	if err != nil {
		logger.Errorw("failed to connect to database", "error", err)

		os.Exit(1)
	}
	logger.Info("connected to database")

	err = migrateDb()
	if err != nil {
		return err
	}

	return nil
}

func migrateDb() error {
	err := db.AutoMigrate(&todos.ToDoItem{})
	if err != nil {
		return err
	}

	return nil
}
