package main

import (
	"flag"
	"log"
	"time"
	"tradedotdotfun-backend/api/controller"
	"tradedotdotfun-backend/api/db"

	"github.com/gofiber/fiber/v2/middleware/pprof"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var (
	addr = flag.String("addr", ":8080", "TCP address to listen to")
)

func main() {
	flag.Parse()
	db.Init()

	app := fiber.New(fiber.Config{
		ReadTimeout: 60 * time.Second,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))
	app.Use(pprof.New(pprof.Config{Prefix: "/a18761"}))

	app.Use("/", func(c *fiber.Ctx) error {
		start := time.Now()
		defer func() {
			end := time.Since(start)

			log.Printf("[AccessLog] %s - %d %d %v - %s\n", getRemoteIP(c), c.Response().StatusCode(), len(c.Response().Body()), end, string(c.Request().RequestURI()))
		}()

		c.Response().Header.SetContentType("application/json")

		return c.Next()
	})

	app.Get("/health", func(ctx *fiber.Ctx) error {
		return nil
	})

	controller.SetPriceRouter(app.Group("/price"))
	controller.SetChartRouter(app.Group("/chart"))
	controller.SetRoundRouter(app.Group("/round"))
	controller.SetPositionRouter(app.Group("/position"))
	controller.SetAccountRouter(app.Group("/account"))
	controller.SetLeaderBoardRouter(app.Group("/leaderboard"))

	log.Println("Ready to Serve")

	if err := app.Listen(*addr); err != nil {
		log.Printf("Listen Error %+v\n", err)
	}
}

func getRemoteIP(ctx *fiber.Ctx) string {
	xForwardedFor := ctx.Request().Header.Peek("x-forwarded-for")
	if xForwardedFor != nil {
		return string(xForwardedFor)
	}

	return ctx.Context().RemoteIP().String()
}
