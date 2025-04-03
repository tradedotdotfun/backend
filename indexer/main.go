package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"tradedotdotfun-backend/common/db"
	"tradedotdotfun-backend/indexer/export"
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

	LeaderBoardManager := leaderboard.NewLeaderBoardManager()
	LeaderBoardManager.Start()

	programExporter := export.NewProgramExporter()
	programExporter.Export()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println("Gracefully shutting down...")
}
