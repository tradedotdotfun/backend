package export

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
	commonconfig "tradedotdotfun-backend/common/config"
	"tradedotdotfun-backend/common/db"
	"tradedotdotfun-backend/common/model"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/config"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProgramExporter struct {
	client *rpc.Client
}

func NewProgramExporter() *ProgramExporter {
	return &ProgramExporter{
		client: rpc.New(commonconfig.RPC_URL),
	}
}

func (e *ProgramExporter) Export() {
	go func() {
		for {
			log.Println("Run ProgramExporter")
			start := time.Now()
			e.FetchProgramData()
			log.Println("Run ProgramExporter Completed, elapsed:", time.Since(start))
			time.Sleep(config.DEPOSIT_UPDATE_INTERVAL)
		}
	}()
}

func (e *ProgramExporter) FetchProgramData() error {
	// 1. 최근에 처리한 트랜잭션을 가져온다.
	db := db.GetConnection()
	var lastSignatureStr string
	var lastSignature *solana.Signature

	db.Model(&model.KeyValueStore{}).
		Select("value").
		Where("key = ?", model.LAST_TX_SIGNATURE_KEY).
		Find(&lastSignatureStr)
	if lastSignatureStr == "" {
		lastSignature = nil
	} else {
		sig := solana.MustSignatureFromBase58(lastSignatureStr)
		lastSignature = &sig
	}

	// 2. 해당 트랜잭션 이전의 모든 트랜잭션을 가져온다.
	rawTxSigs, err := e.getSignaturesForAddress(lastSignature)
	if err != nil {
		return errors.New("failed to get signatures for address")
	}

	// 3. 트랜잭션이 최신 상태부터 정렬되어 있기 때문에 역순으로 정렬한다.
	reversedTxSigs := make([]*rpc.TransactionSignature, len(rawTxSigs))
	for i, sig := range rawTxSigs {
		reversedTxSigs[len(rawTxSigs)-1-i] = sig
	}

	// 과거 순으로 정렬되어있는지 확인한다.
	for i := 1; i < len(reversedTxSigs); i++ {
		if reversedTxSigs[i-1].Slot > reversedTxSigs[i].Slot {
			return errors.New("reversedTxSigs is not sorted by slot in ascending order")
		}
	}

	// 4. DepositEvent를 추출하고 account를 생성한다.
	for _, sig := range reversedTxSigs {
		tx, err := e.getTransaction(sig.Signature)
		if err != nil {
			return errors.New("failed to get transaction")
		}

		if exist, round := e.extractRoundStartEvent(tx); exist {
			if err := e.saveStartRound(round, sig.Signature); err != nil {
				return errors.New("failed to save start round")
			}
		}

		if exist, round := e.extractRoundEndEvent(tx); exist {
			if err := e.saveEndRound(round, sig.Signature); err != nil {
				return errors.New("failed to save end round")
			}
		}

		if exist, round := e.extractDistributeEvent(tx); exist {
			if err := e.saveDistribute(round, sig.Signature); err != nil {
				return errors.New("failed to save distribute round")
			}
		}

		if exist, address, round := e.extractParticipatedEvent(tx); exist {
			if err := e.saveDepositAndCreateAccount(address, round, sig); err != nil {
				return errors.New("failed to save deposit and create account")
			}
		}

	}

	return nil
}

func (e *ProgramExporter) getSignaturesForAddress(lastSignature *solana.Signature) ([]*rpc.TransactionSignature, error) {
	ctx := context.Background()
	limit := 1000
	options := &rpc.GetSignaturesForAddressOpts{Limit: &limit}
	if lastSignature != nil {
		options.Until = *lastSignature
	}
	return e.client.GetSignaturesForAddressWithOpts(ctx, commonconfig.DEPOSIT_ADDRESS, options)
}

func (e *ProgramExporter) getTransaction(txSig solana.Signature) (*rpc.GetTransactionResult, error) {
	ctx := context.Background()
	maxSupplortedTransactionVersion := uint64(2)
	return e.client.GetTransaction(ctx, txSig, &rpc.GetTransactionOpts{
		MaxSupportedTransactionVersion: &maxSupplortedTransactionVersion,
	})
}

func (e *ProgramExporter) filterFinalizedSignatures(txSigs []*rpc.TransactionSignature) []*rpc.TransactionSignature {
	var finalizedSignatures []*rpc.TransactionSignature
	for _, sig := range txSigs {
		if sig.ConfirmationStatus == "finalized" {
			finalizedSignatures = append(finalizedSignatures, sig)
		}
	}
	return finalizedSignatures
}

// contract를 통해서 deposit 하는 경우는 무시한다.
/*
	Example
	[
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj invoke [1]",
		"Program log: Instruction: DepositSol",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program log: User participated: user=2qKn8e94V8HiemUQJ3vG13zC555TBbxFkmbqgU4tUDXJ round=1",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj consumed 18386 of 200000 compute units",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj success"
	]
*/
func (e *ProgramExporter) extractParticipatedEvent(tx *rpc.GetTransactionResult) (bool, string, uint64) {
	logLength := 8
	txLogs := tx.Meta.LogMessages
	logPattern := regexp.MustCompile(`Program log: User participated: user=([^ ]+) round=(\d+)`)
	// txLogs 배열이 정확한 순서로 되어 있는지 확인하고, 조건이 맞으면 주소 추출
	for i := 0; i <= len(txLogs)-logLength; i++ {
		if txLogs[i] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" invoke [1]" &&
			txLogs[i+1] == "Program log: Instruction: DepositSol" &&
			txLogs[i+2] == "Program 11111111111111111111111111111111 invoke [2]" &&
			txLogs[i+3] == "Program 11111111111111111111111111111111 success" &&
			strings.HasPrefix(txLogs[i+5], "Program data:") &&
			strings.HasPrefix(txLogs[i+6], "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" consumed") &&
			txLogs[i+7] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" success" {

			matches4 := logPattern.FindStringSubmatch(txLogs[i+4])

			if len(matches4) > 2 {
				userPubkey := matches4[1]
				round, err := strconv.ParseUint(matches4[2], 10, 64)
				if err != nil {
					log.Println("Error parsing round:", err)
					return false, "", 0
				}
				return true, userPubkey, round
			}
		}
	}
	return false, "", 0
}

// contract를 통해서 round start 하는 경우는 무시한다.
/*
	Example
	[
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj invoke [1]",
		"Program log: Instruction: StartRound",
		"Program log: Round started: 1",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj consumed 3694 of 200000 compute units",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj success"
	]
*/
func (e *ProgramExporter) extractRoundStartEvent(tx *rpc.GetTransactionResult) (bool, uint64) {
	logLength := 5
	txLogs := tx.Meta.LogMessages
	logPattern := regexp.MustCompile(`Program log: Round started: (\d+)`)
	// txLogs 배열이 정확한 순서로 되어 있는지 확인하고, 조건이 맞으면 주소 추출
	for i := 0; i <= len(txLogs)-logLength; i++ {
		if txLogs[i] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" invoke [1]" &&
			txLogs[i+1] == "Program log: Instruction: StartRound" &&
			strings.HasPrefix(txLogs[i+3], "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" consumed") &&
			txLogs[i+4] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" success" {

			matches2 := logPattern.FindStringSubmatch(txLogs[i+2])

			if len(matches2) > 1 {
				round, err := strconv.ParseUint(matches2[1], 10, 64)
				if err != nil {
					log.Println("Error parsing round:", err)
					return false, 0
				}
				return true, round
			}
		}
	}
	return false, 0
}

// contract를 통해서 round end 하는 경우는 무시한다.
/*
	Example
	[
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj invoke [1]",
		"Program log: Instruction: EndRound",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program log: Round ended: 1",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj consumed 7995 of 200000 compute units",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj success"
	]
*/
func (e *ProgramExporter) extractRoundEndEvent(tx *rpc.GetTransactionResult) (bool, uint64) {
	logLength := 7
	txLogs := tx.Meta.LogMessages
	logPattern := regexp.MustCompile(`Program log: Round ended: (\d+)`)
	// txLogs 배열이 정확한 순서로 되어 있는지 확인하고, 조건이 맞으면 주소 추출
	for i := 0; i <= len(txLogs)-logLength; i++ {
		if txLogs[i] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" invoke [1]" &&
			txLogs[i+1] == "Program log: Instruction: EndRound" &&
			txLogs[i+2] == "Program 11111111111111111111111111111111 invoke [2]" &&
			txLogs[i+3] == "Program 11111111111111111111111111111111 success" &&
			strings.HasPrefix(txLogs[i+5], "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" consumed") &&
			txLogs[i+6] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" success" {

			matches2 := logPattern.FindStringSubmatch(txLogs[i+4])

			if len(matches2) > 1 {
				round, err := strconv.ParseUint(matches2[1], 10, 64)
				if err != nil {
					log.Println("Error parsing round:", err)
					return false, 0
				}
				return true, round
			}
		}
	}
	return false, 0
}

// contract를 통해서 distribute 하는 경우는 무시한다.
/*
	Example
	[
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj invoke [1]",
		"Program log: Instruction: DistributeSol",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program log: SOL distributed for Round 2",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj consumed 8411 of 200000 compute units",
		"Program ANE115teNkEVcqpKF7F9tiBYj89o5mXqhXgMBkNaBtbj success"
	]
*/
func (e *ProgramExporter) extractDistributeEvent(tx *rpc.GetTransactionResult) (bool, uint64) {
	logLength := 7
	txLogs := tx.Meta.LogMessages
	logPattern := regexp.MustCompile(`Program log: SOL distributed for Round (\d+)`)
	// txLogs 배열이 정확한 순서로 되어 있는지 확인하고, 조건이 맞으면 주소 추출
	for i := 0; i <= len(txLogs)-logLength; i++ {
		if txLogs[i] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" invoke [1]" &&
			txLogs[i+1] == "Program log: Instruction: DistributeSol" &&
			txLogs[i+2] == "Program 11111111111111111111111111111111 invoke [2]" &&
			txLogs[i+3] == "Program 11111111111111111111111111111111 success" &&
			strings.HasPrefix(txLogs[i+5], "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" consumed") &&
			txLogs[i+6] == "Program "+commonconfig.DEPOSIT_ADDRESS.String()+" success" {

			matches2 := logPattern.FindStringSubmatch(txLogs[i+4])

			if len(matches2) > 1 {
				round, err := strconv.ParseUint(matches2[1], 10, 64)
				if err != nil {
					log.Println("Error parsing round:", err)
					return false, 0
				}
				return true, round
			}
		}
	}
	return false, 0
}

func (e *ProgramExporter) saveStartRound(round uint64, txSig solana.Signature) error {
	db := db.GetConnection()
	lastRound := cache.GetRound()
	lastRoundStatus := cache.GetRoundStatus()

	// 일단은 로그만 남기자
	if lastRound+1 != round {
		log.Println("Round mismatch. Expected:", lastRound+1, "Got:", round)
	}
	if lastRoundStatus != model.RoundStatusReady {
		log.Println("Round status mismatch. Expected:", model.RoundStatusReady, "Got:", lastRoundStatus)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		keyValueStores := []model.KeyValueStore{
			{
				Key:   model.CURRENT_ROUND_KEY,
				Value: strconv.FormatInt(int64(round), 10),
			},
			{
				Key:   model.CURRENT_ROUND_STATUS_KEY,
				Value: string(model.RoundStatusStart),
			},
			{
				Key:   model.LAST_TX_SIGNATURE_KEY,
				Value: txSig.String(),
			},
		}
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(keyValueStores).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	cache.SetRound(round)
	cache.SetRoundStatus(model.RoundStatusStart)
	return nil
}

func (e *ProgramExporter) saveEndRound(round uint64, txSig solana.Signature) error {
	db := db.GetConnection()
	lastRound := cache.GetRound()
	lastRoundStatus := cache.GetRoundStatus()

	// 일단은 로그만 남기자
	if lastRound != round {
		log.Println("Round mismatch. Expected:", lastRound, "Got:", round)
	}
	if lastRoundStatus != model.RoundStatusStart {
		log.Println("Round status mismatch. Expected:", model.RoundStatusStart, "Got:", lastRoundStatus)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		keyValueStores := []model.KeyValueStore{
			{
				Key:   model.CURRENT_ROUND_KEY,
				Value: strconv.FormatInt(int64(round), 10),
			},
			{
				Key:   model.CURRENT_ROUND_STATUS_KEY,
				Value: string(model.RoundStatusEnd),
			},
			{
				Key:   model.LAST_TX_SIGNATURE_KEY,
				Value: txSig.String(),
			},
		}
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(keyValueStores).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	cache.SetRound(round)
	cache.SetRoundStatus(model.RoundStatusEnd)
	return nil
}

func (e *ProgramExporter) saveDistribute(round uint64, txSig solana.Signature) error {
	db := db.GetConnection()
	lastRound := cache.GetRound()
	lastRoundStatus := cache.GetRoundStatus()

	// 일단은 로그만 남기자
	if lastRound != round {
		log.Println("Round mismatch. Expected:", lastRound, "Got:", round)
	}
	if lastRoundStatus != model.RoundStatusEnd {
		log.Println("Round status mismatch. Expected:", model.RoundStatusEnd, "Got:", lastRoundStatus)
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		keyValueStores := []model.KeyValueStore{
			{
				Key:   model.CURRENT_ROUND_KEY,
				Value: strconv.FormatInt(int64(round), 10),
			},
			{
				Key:   model.CURRENT_ROUND_STATUS_KEY,
				Value: string(model.RoundStatusReady),
			},
			{
				Key:   model.LAST_TX_SIGNATURE_KEY,
				Value: txSig.String(),
			},
		}
		if err := tx.Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{"value"}),
		}).Create(keyValueStores).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	cache.SetRound(round)
	cache.SetRoundStatus(model.RoundStatusReady)
	return nil
}

func (e *ProgramExporter) saveDepositAndCreateAccount(address string, round uint64, txSig *rpc.TransactionSignature) error {
	db := db.GetConnection()
	currentRound := cache.GetRound()
	currentRoundStatus := cache.GetRoundStatus()

	if round != currentRound {
		log.Println("Round mismatch. Expected:", currentRound, "Got:", round)
		return nil
	}
	if currentRoundStatus != model.RoundStatusStart {
		log.Println("Round status mismatch. Expected:", model.RoundStatusStart, "Got:", currentRoundStatus)
		return nil
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// round와 address가 같고 deleted가 false인 account가 이미 있다면 deleted를 true로 업데이트
		if err := tx.Model(&model.Account{}).
			Where("round = ? AND address = ? AND deleted = ?", round, address, false).
			Update("deleted", true).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Position{}).
			Where("round = ? AND address = ? AND status = ?", round, address, model.StatusActive).
			Update("status", model.StatusRemoved).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.KeyValueStore{}).
			Where("key = ?", model.LAST_TX_SIGNATURE_KEY).
			Update("value", txSig.Signature.String()).Error; err != nil {
			return err
		}

		// 새로운 account 생성
		if err := tx.Create(&model.Account{
			Round:     round,
			Address:   address,
			USDAmount: config.USER_USD_AMOUNT,
		}).Error; err != nil {
			return err
		}

		// depositEvent 생성
		if err := tx.Create(&model.DepositEvent{
			Signature: txSig.Signature.String(),
			Slot:      txSig.Slot,
			Address:   address,
		}).Error; err != nil {
			return err
		}

		return nil
	})
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		// 이미 존재하는 데이터는 무시한다.
		return nil
	}
	return err
}
