package controller

import (
	"tradedotdotfun-backend/api/service"
	"tradedotdotfun-backend/common/utils"

	"github.com/gofiber/fiber/v2"
)

func SetRoundRouter(r fiber.Router) {
	service.StartRoundUpdater()
	r.Get("/", utils.Wrap(HandleGetRound))
}

func HandleGetRound(c *fiber.Ctx) (interface{}, error) {
	return service.GetRound(), nil
}
