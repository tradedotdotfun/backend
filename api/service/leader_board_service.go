package service

import (
	"log"
	"strconv"
	"time"
	"tradedotdotfun-backend/api/db"
	"tradedotdotfun-backend/api/types"

	"github.com/patrickmn/go-cache"
)

var leaderboardCache = cache.New(10*time.Minute, 10*time.Minute)

func GetLeaderBoard(round uint64, limit uint64) types.GetLeaderBoardResponse {
	leaderBoard, found := leaderboardCache.Get(strconv.FormatUint(round, 10))
	if found {
		return applyLimit(leaderBoard.([]types.LeaderBoardData), limit)
	}
	leaderBoard, err := FetchLeaderBoardFromDB(round)
	if err != nil {
		log.Println(err)
		return types.GetLeaderBoardResponse{}
	}
	leaderboardCache.Set(strconv.FormatUint(round, 10), leaderBoard, cache.DefaultExpiration)
	return applyLimit(leaderBoard.([]types.LeaderBoardData), limit)
}

func FetchLeaderBoardFromDB(round uint64) ([]types.LeaderBoardData, error) {
	conn := db.GetConnection()

	var leaderBoardData []types.LeaderBoardData
	conn.Table("leader_boards").
		Select("leader_boards.rank, leader_boards.address, address_names.name AS name, leader_boards.pnl, leader_boards.roi").
		Joins("LEFT JOIN address_names ON leader_boards.address = address_names.address").
		Where("round = ?", round).
		Order("rank ASC").
		Find(&leaderBoardData)

	return leaderBoardData, nil
}

func applyLimit(leaderBoard []types.LeaderBoardData, limit uint64) []types.LeaderBoardData {
	if limit == 0 || limit > uint64(len(leaderBoard)) {
		limit = uint64(len(leaderBoard))
	}
	return leaderBoard[:limit]
}
