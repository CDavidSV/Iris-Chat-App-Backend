package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CDavidSV/Iris-Chat-App-Backend/cmd/api/handlers"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func connectDB(connString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func main() {
	godotenv.Load()

	fiberConfig := fiber.Config{
		ServerHeader: "IrisAPI_v1",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	app := fiber.New(fiberConfig)

	pool, err := connectDB(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	fmt.Println("Successfully connected to postgres!")
	defer pool.Close()

	// setup logger
	app.Use(logger.New(logger.Config{
		TimeFormat: "02/06/2006 15:04:05",
	}))

	// cors config
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowMethods: "POST, GET, OPTION, PUT, DELETE, HEAD",
	}))

	websocketServer := websocket.WebsocketServer{
		DB:                pool,
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
	}

	// websockets
	app.Use("/ws", websocketServer.WebsocketUpgrade)
	app.Get("/ws", websocketServer.NewWebsocket())

	server := &handlers.Server{
		DBpool:        pool,
		Users:         &models.UserModel{DB: pool},
		Sessions:      &models.SessionsModel{DB: pool},
		Relationships: &models.RelationshipModel{DB: pool},
	}

	// load routes
	server.LoadRoutes(app)

	app.Listen(":3000")
}
