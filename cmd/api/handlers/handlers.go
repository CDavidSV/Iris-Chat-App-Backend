package handlers

import (
	"github.com/CDavidSV/Iris-Chat-App-Backend/cmd/api/middleware"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DBpool        *pgxpool.Pool
	Users         *models.UserModel
	Sessions      *models.SessionsModel
	Relationships *models.RelationshipModel

	Websocket *websocket.WebsocketServer
}

func (s *Server) LoadRoutes(app *fiber.App) {
	// ------------------ Unprotected routes ------------------
	app.Post("/auth/signup", s.Register)
	app.Post("/auth/login", s.Login)

	// ------------------ Protected routes ------------------

	// Auth
	app.Post("/auth/logout", middleware.Authorize, s.Logout)
	app.Post("/auth/token", s.Token)

	// User
	app.Get("user/me", middleware.Authorize, s.GetMe)
	app.Get("user/:userID", middleware.Authorize, s.GetUser)

	// Relationships
	app.Get("/relationships", middleware.Authorize, s.GetRelationships)
	app.Post("/relationships/:username", middleware.Authorize, s.CreateRelationship)
	app.Delete("/relationships/:userID", middleware.Authorize, s.DelRelationship)
}
