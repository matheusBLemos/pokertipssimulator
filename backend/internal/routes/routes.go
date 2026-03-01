package routes

import (
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"pokertipssimulator/internal/handler"
)

func Setup(app *fiber.App, roomH *handler.RoomHandler, gameH *handler.GameHandler, actionH *handler.ActionHandler, wsH *handler.WSHandler, jwtSecret string) {
	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	// Public routes
	api.Post("/rooms", roomH.CreateRoom)
	api.Post("/rooms/join", roomH.JoinRoom)

	// Auth middleware
	auth := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing authorization header")
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Locals("roomID", claims["room_id"])
		c.Locals("playerID", claims["player_id"])
		c.Locals("isHost", claims["is_host"])

		return c.Next()
	}

	// Protected routes
	rooms := api.Group("/rooms/:roomId", auth)
	rooms.Get("/", roomH.GetRoom)
	rooms.Put("/config", roomH.UpdateConfig)
	rooms.Put("/players/:playerId/seat", roomH.PickSeat)
	rooms.Post("/rounds/start", gameH.StartRound)
	rooms.Post("/rounds/advance", gameH.AdvanceStreet)
	rooms.Post("/rounds/settle", gameH.SettleRound)
	rooms.Post("/pause", gameH.PauseGame)
	rooms.Post("/players/:playerId/rebuy", gameH.Rebuy)
	rooms.Post("/action", actionH.PerformAction)
	rooms.Delete("/players/:playerId", gameH.KickPlayer)

	// WebSocket
	app.Use("/ws", wsH.Upgrade)
	app.Get("/ws", websocket.New(wsH.Handle))
}
