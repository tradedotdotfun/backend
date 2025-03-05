package controller

import (
	"errors"
	"tradedotdotfun-backend/api/service"
	"tradedotdotfun-backend/common/utils"

	"github.com/gofiber/fiber/v2"
)

func SetChartRouter(r fiber.Router) {
	service.StartChartUpdater()
	r.Get("/", utils.Wrap(HandleChart))
}

func HandleChart(c *fiber.Ctx) (interface{}, error) {
	coinID := c.Query("coin_id")
	if coinID == "" {
		return nil, errors.New("coin_id is required")
	}
	return service.GetChart(coinID), nil
}
