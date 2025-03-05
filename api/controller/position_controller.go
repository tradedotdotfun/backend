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

func SetPositionRouter(r fiber.Router) {
	r.Post("/", utils.Wrap(HandleCreatePosition))
	r.Post("/:position_id/close", utils.Wrap(HandleClosePosition))
	r.Get("/", utils.Wrap(HandleGetPosition))
}

func HandleCreatePosition(c *fiber.Ctx) (interface{}, error) {
	address, err := auth.CheckAuthorization(c)
	if err != nil {
		return nil, err
	}

	var body types.CreatePositionRequest
	if err := c.BodyParser(&body); err != nil {
		return nil, err
	}

	if err := body.Validate(); err != nil {
		return nil, err
	}

	return service.CreatePosition(address, &body)
}

func HandleClosePosition(c *fiber.Ctx) (interface{}, error) {
	address, err := auth.CheckAuthorization(c)
	if err != nil {
		return nil, err
	}

	positionIdStr := c.Params("position_id")

	positionId, err := strconv.ParseUint(positionIdStr, 10, 64)
	if err != nil {
		return nil, err
	}

	var body types.ClosePositionRequest
	if err := c.BodyParser(&body); err != nil {
		return nil, err
	}

	if err := body.Validate(); err != nil {
		return nil, err
	}

	return service.ClosePosition(address, positionId, body.Percentage)
}

func HandleGetPosition(c *fiber.Ctx) (interface{}, error) {
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
	return service.GetPosition(round, address), nil
}
