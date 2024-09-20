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
	Channels      *models.ChannelModel

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
	app.Get("users/me", middleware.Authorize, s.GetMe)
	app.Get("users/username/:username", middleware.Authorize, s.GetUsersByUsername)
	app.Get("users/id/:userID", middleware.Authorize, s.GetUser)

	// Relationships
	app.Get("/relationships/friends", middleware.Authorize, s.GetFriendships)
	app.Get("/relationships/requests", middleware.Authorize, s.GetFriendRequests)
	app.Get("/relationships/blocked", middleware.Authorize, s.GetBlockedUsers)
	app.Post("/relationships/:userID", middleware.Authorize, s.CreateRelationship)
	app.Delete("/relationships/:userID", middleware.Authorize, s.DelRelationship)
	app.Put("/relationships/:userID", middleware.Authorize, s.BlockUser)

	// Profile
	app.Put("/profile/update", middleware.Authorize, s.UpdateProfile)
	app.Post("/profile/change-password", middleware.Authorize, s.ChangePassword)
}
