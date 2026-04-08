package handler

import (
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func SetupRoutes(
	app *fiber.App,
	roomH *RoomHandler,
	gameH *GameHandler,
	actionH *ActionHandler,
	tipsH *TipsHandler,
	wsH *WSHandler,
	jwtSecret string,
) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	auth := jwtMiddleware(jwtSecret)

	api := app.Group("/api/v1")

	// --- Game mode routes ---
	gameAPI := api.Group("/game")
	gameAPI.Post("/rooms", roomH.CreateRoom)
	gameAPI.Post("/rooms/join", roomH.JoinRoom)

	gameRooms := gameAPI.Group("/rooms/:roomId", auth)
	gameRooms.Get("/", roomH.GetRoom)
	gameRooms.Put("/config", roomH.UpdateConfig)
	gameRooms.Put("/players/:playerId/seat", roomH.PickSeat)
	gameRooms.Post("/rounds/start", gameH.StartRound)
	gameRooms.Post("/rounds/advance", gameH.AdvanceStreet)
	gameRooms.Post("/rounds/settle", gameH.SettleRound)
	gameRooms.Post("/rounds/auto-settle", gameH.AutoSettleRound)
	gameRooms.Post("/pause", gameH.PauseGame)
	gameRooms.Post("/players/:playerId/rebuy", gameH.Rebuy)
	gameRooms.Post("/action", actionH.PerformAction)
	gameRooms.Delete("/players/:playerId", gameH.KickPlayer)

	// --- Tips mode routes ---
	tipsAPI := api.Group("/tips")
	tipsAPI.Post("/rooms", roomH.CreateRoom)
	tipsAPI.Post("/rooms/join", roomH.JoinRoom)

	tipsRooms := tipsAPI.Group("/rooms/:roomId", auth)
	tipsRooms.Get("/", roomH.GetRoom)
	tipsRooms.Put("/config", roomH.UpdateConfig)
	tipsRooms.Put("/players/:playerId/seat", roomH.PickSeat)
	tipsRooms.Post("/chips/transfer", tipsH.TransferChips)
	tipsRooms.Post("/blinds/advance", tipsH.AdvanceBlind)
	tipsRooms.Post("/pause", tipsH.PauseTimer)
	tipsRooms.Post("/players/:playerId/rebuy", tipsH.Rebuy)
	tipsRooms.Delete("/players/:playerId", tipsH.KickPlayer)

	// --- WebSocket (shared) ---
	app.Use("/ws", wsH.Upgrade)
	app.Get("/ws", websocket.New(wsH.Handle))
}

func jwtMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
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
}
