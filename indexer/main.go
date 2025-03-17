package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"tradedotdotfun-backend/indexer/db"
	"tradedotdotfun-backend/indexer/deposit"
	"tradedotdotfun-backend/indexer/leaderboard"
	"tradedotdotfun-backend/indexer/okx"
	"tradedotdotfun-backend/indexer/round"
)

func main() {
	flag.Parse()
	db.Init()

	okxFetcher := okx.NewFetcher()
	okxFetcher.Fetch()

	roundManager := round.NewRoundManager()
	roundManager.Start()

	leaderBoardProcessor := leaderboard.NewLeaderBoardProcessor()
	leaderBoardProcessor.Start()

	depositExporter := deposit.NewDepositExporter()
	depositExporter.Export()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Gracefully shutting down...")
}
