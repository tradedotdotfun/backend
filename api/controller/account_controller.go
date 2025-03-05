package controller

import (
	"errors"
	"strconv"
	"tradedotdotfun-backend/api/auth"
	"tradedotdotfun-backend/api/service"
	"tradedotdotfun-backend/api/types"
	"tradedotdotfun-backend/common/utils"

	"github.com/gofiber/fiber/v2"
)

func SetAccountRouter(r fiber.Router) {
	r.Get("/", utils.Wrap(HandleGetAccount))
	r.Post("/name", utils.Wrap(HandleAddName))
}

func HandleGetAccount(c *fiber.Ctx) (interface{}, error) {
	roundStr := c.Query("round")
	if roundStr == "" {
		return nil, errors.New("round is required")
	}
	address := c.Query("address")
	if address == "" {
		return nil, errors.New("address is required")
	}

	round, err := strconv.ParseUint(roundStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return service.GetAccount(round, address), nil
}

func HandleAddName(c *fiber.Ctx) (interface{}, error) {
	address, err := auth.CheckAuthorization(c)
	if err != nil {
		return nil, err
	}

	var body types.AddNameRequest
	if err := c.BodyParser(&body); err != nil {
		return nil, err
	}

	return service.AddName(address, body.Name)
}
