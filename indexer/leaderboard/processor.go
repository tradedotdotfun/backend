package leaderboard

import (
	"fmt"
	"time"
	"tradedotdotfun-backend/common/service"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/config"

	"github.com/robfig/cron/v3"
)

type LeaderBoardManager struct {
	leaderBoardProcessor *service.LeaderBoardProcessor
	cron                 *cron.Cron
}

func NewLeaderBoardManager() *LeaderBoardManager {
	return &LeaderBoardManager{
		leaderBoardProcessor: service.NewLeaderBoardProcessor(),
		cron:                 cron.New(cron.WithSeconds()), // 초 단위 스케줄러
	}
}

func (lbm *LeaderBoardManager) Start() {
	lbm.process()
	_, err := lbm.cron.AddFunc(config.LEADER_BOARD_CRON_SPEC, lbm.process)
	if err != nil {
		fmt.Println("Cron job 추가 중 오류 발생:", err)
		return
	}

	lbm.cron.Start()
}

func (lbp *LeaderBoardManager) process() {
	fmt.Println("리더보드 프로세서 시작! 현재 시간:", time.Now().Format("2006-01-02 15:04:05"))
	defer fmt.Println("리더보드 프로세서 종료! 현재 시간:", time.Now().Format("2006-01-02 15:04:05"))
	round := cache.GetRound()
	price := cache.GetPrice()
	lbp.leaderBoardProcessor.Process(round, price)
}
