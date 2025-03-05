package controller

import (
	"errors"
	"strconv"
	"tradedotdotfun-backend/api/service"
	"tradedotdotfun-backend/common/utils"

	"github.com/gofiber/fiber/v2"
)

func SetLeaderBoardRouter(r fiber.Router) {
	r.Get("/", utils.Wrap(HandleGetLeaderBoard))
}

func HandleGetLeaderBoard(c *fiber.Ctx) (interface{}, error) {
	roundStr := c.Query("round")
	if roundStr == "" {
		return nil, errors.New("round is required")
	}
	round, err := strconv.ParseUint(roundStr, 10, 64)
	if err != nil {
		return nil, err
	}
	limitStr := c.Query("limit")
	limit, err := strconv.ParseUint(limitStr, 10, 64)
	if err != nil {
		limit = 0
	}

	return service.GetLeaderBoard(round, limit), nil
}
