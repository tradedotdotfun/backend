package leaderboard

import (
	"fmt"
	"sort"
	"time"
	commonconfig "tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/model"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/config"
	"tradedotdotfun-backend/indexer/db"

	"github.com/gofiber/fiber/v2/log"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type LeaderBoardProcessor struct {
	cron *cron.Cron
}

func NewLeaderBoardProcessor() *LeaderBoardProcessor {
	return &LeaderBoardProcessor{
		cron: cron.New(cron.WithSeconds()), // 초 단위 스케줄러
	}
}

func (lbp *LeaderBoardProcessor) Start() {
	lbp.process()
	_, err := lbp.cron.AddFunc(config.LEADER_BOARD_CRON_SPEC, lbp.process)
	if err != nil {
		fmt.Println("Cron job 추가 중 오류 발생:", err)
		return
	}

	lbp.cron.Start()
}

func (lbp *LeaderBoardProcessor) process() {
	fmt.Println("리더보드 갱신! 현재 시간:", time.Now().Format("2006-01-02 15:04:05"))
	db := db.GetConnection()
	round := cache.GetRound()
	price := cache.GetPrice()

	var leaderBoards []model.LeaderBoard

	batchSize := 10 // 한 번에 처리할 address 개수

	// 1. 모든 address를 한 번에 가져오기 (전체 트랜잭션 X)
	var addresses []string
	if err := db.Model(&model.Account{}).
		Where("round = ?", round).
		Select("DISTINCT address").
		Pluck("address", &addresses).Error; err != nil {
		log.Errorf("Failed to get addresses: %v", err)
		return
	}

	// 2. address 목록을 batchSize 만큼 나눠서 개별 트랜잭션 처리
	for i := 0; i < len(addresses); i += batchSize {
		end := i + batchSize
		if end > len(addresses) {
			end = len(addresses) // 마지막 batch 처리
		}
		batch := addresses[i:end] // 현재 batch에 해당하는 address 목록

		var batchPositions []model.Position
		var batchAccounts []model.Account
		// 개별 트랜잭션 실행
		err := db.Transaction(func(tx *gorm.DB) error {
			// 3. positions 조회
			if err := tx.Where("round = ? AND address IN (?) AND status = ?", round, batch, model.StatusActive).
				Find(&batchPositions).Error; err != nil {
				return err
			}

			// 4. accounts 조회
			if err := tx.Where("round = ? AND address IN (?) AND deleted = false", round, batch).
				Find(&batchAccounts).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Errorf("Failed to get positions and accounts: %v", err)
			return
		}

		// 5. positions를 address 별로 그룹핑
		positionMap := make(map[string][]model.Position)
		for _, position := range batchPositions {
			positionMap[position.Address] = append(positionMap[position.Address], position)
		}

		// 6. address 별로 leader board 계산
		for _, account := range batchAccounts {
			positions := positionMap[account.Address]

			leaderBoards = append(leaderBoards, calculateLeaderBoard(account, positions, price))
		}
	}

	// 7. 리더보드 정렬 및 Rank 할당
	// PnL 기준 내림차순 정렬
	sort.Slice(leaderBoards, func(i, j int) bool {
		return leaderBoards[i].PnL > leaderBoards[j].PnL
	})
	// Rank 할당
	rank := uint64(1)
	for i := range leaderBoards {
		if i > 0 && leaderBoards[i].PnL == leaderBoards[i-1].PnL {
			leaderBoards[i].Rank = leaderBoards[i-1].Rank
		} else {
			leaderBoards[i].Rank = rank
		}
		rank++
	}

	// 8. 리더보드 저장
	err := db.Transaction(func(tx *gorm.DB) error {
		// round에 있는 기존 리더보드 데이터 삭제
		if err := tx.Where("round = ?", round).Delete(&model.LeaderBoard{}).Error; err != nil {
			return err
		}

		// 새로운 리더보드 데이터 배치 저장
		if err := tx.Create(&leaderBoards).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Errorf("Failed to save leader boards transactionally: %v", err)
		return
	}
}

func calculateLeaderBoard(account model.Account, positions []model.Position, price map[string]float64) model.LeaderBoard {
	var totalPnL float64
	var totalAmount float64
	for _, position := range positions {
		currentPrice := price[position.Token]
		currentValue := currentPrice * position.PositionSize
		initialAmount := position.Amount * float64(position.Leverage)
		var pnl float64
		if position.Type == "long" {
			pnl = currentValue - initialAmount
		} else {
			pnl = initialAmount - currentValue
		}

		if position.Amount+pnl <= 0 {
			continue
		}

		totalAmount += position.Amount
		totalPnL += pnl
	}
	totalAmount = totalAmount + account.USDAmount
	totalPnL += (totalAmount - commonconfig.INITIAL_USD_BALANCE)
	leaderBoard := model.LeaderBoard{
		Round:   account.Round,
		Address: account.Address,
		PnL:     totalPnL,
		RoI:     totalPnL / commonconfig.INITIAL_USD_BALANCE,
	}
	return leaderBoard
}
