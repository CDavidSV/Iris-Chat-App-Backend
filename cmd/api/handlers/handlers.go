package handlers

import (
	"github.com/CDavidSV/Iris-Chat-App-Backend/cmd/api/middleware"
	"github.com/CDavidSV/Iris-Chat-App-Backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	DBpool   *pgxpool.Pool
	Users    *models.UserModel
	Sessions *models.SessionsModel
}

func (s *Server) LoadRoutes(app *fiber.App) {
	// Unprotected routes
	app.Post("/auth/signup", s.Register)
	app.Post("/auth/login", s.Login)

	// Protected routes
	app.Post("/auth/logout", middleware.Authorize, s.Logout)
	app.Post("/auth/token", s.Token)
}
