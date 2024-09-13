package handlers

import (
	"github.com/CDavidSV/Iris-Chat-App-Backend/cmd/api/middleware"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DBpool        *pgxpool.Pool
	Users         *models.UserModel
	Sessions      *models.SessionsModel
	Relationships *models.RelationshipModel
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
}
