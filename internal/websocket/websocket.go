package websocket

import (
	"log"
	"net/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebsocketClient struct {
	Status string
	Conn   *websocket.Conn
}

type WebsocketServer struct {
	Connections       map[string]WebsocketClient
	AccessTokenSecret string
	DB                *pgxpool.Pool
}

func (ws *WebsocketServer) WebsocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		// Validate access token
		accessToken := c.Query("token")
		if accessToken == "" {
			return c.SendStatus(http.StatusUnauthorized)
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(ws.AccessTokenSecret), nil
		})
		if err != nil {
			return c.SendStatus(http.StatusUnauthorized)
		}
		if !token.Valid {
			return c.SendStatus(http.StatusUnauthorized)
		}

		c.Locals("userID", claims["userID"])
		c.Locals("sessionID", claims["sessionID"])

		return c.Next()
	}

	return fiber.ErrUpgradeRequired
}

func (ws *WebsocketServer) NewWebsocket() func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		if len(ws.Connections) == 0 {
			ws.Connections = make(map[string]WebsocketClient)
		}

		// Add the connection to the map
		ws.Connections[c.Locals("userID").(string)] = WebsocketClient{
			Status: "online",
			Conn:   c,
		}

		ws.readLoop(c, c.Locals("userID").(string)) // listen for messages
	})
}

func (ws *WebsocketServer) readLoop(conn *websocket.Conn, userID string) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(ws.Connections, userID)
			return
		}
	}
}
