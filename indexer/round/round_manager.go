package round

import (
	"fmt"
	"time"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/config"

	"github.com/robfig/cron/v3"
)

// RoundManager 구조체 정의
type RoundManager struct {
	cron *cron.Cron
}

// NewRoundManager: RoundManager 인스턴스 생성
func NewRoundManager() *RoundManager {
	// TODO: 임시 처리
	cache.SetRound(0)
	return &RoundManager{
		cron: cron.New(cron.WithSeconds()), // 초 단위 스케줄러
	}
}

// Start: 매일 00:00에 실행되도록 설정
func (rm *RoundManager) Start() {
	_, err := rm.cron.AddFunc(config.ROUND_CRON_SPEC, rm.runRound)
	if err != nil {
		fmt.Println("Cron job 추가 중 오류 발생:", err)
		return
	}

	rm.cron.Start()
}

// runRound: 실제 실행할 작업
func (rm *RoundManager) runRound() {
	fmt.Println("새로운 라운드 시작! 현재 시간:", time.Now().Format("2006-01-02 15:04:05"))
}
