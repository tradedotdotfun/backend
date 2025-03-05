package deposit

import (
	"context"
	"errors"
	"log"
	"regexp"
	"strings"
	"time"
	"tradedotdotfun-backend/common/model"
	"tradedotdotfun-backend/indexer/cache"
	"tradedotdotfun-backend/indexer/config"
	"tradedotdotfun-backend/indexer/db"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"gorm.io/gorm"
)

type DepositExporter struct {
	client *rpc.Client
}

func NewDepositExporter() *DepositExporter {
	return &DepositExporter{
		client: rpc.New(config.SOLANA_RPC_URL),
	}
}

func (e *DepositExporter) Export() {
	go func() {
		for {
			log.Println("Run DepositExporter")
			start := time.Now()
			e.FetchDepositData()
			log.Println("Run DepositExporter Completed, elapsed:", time.Since(start))
			time.Sleep(config.DEPOSIT_UPDATE_INTERVAL)
		}
	}()
}

func (e *DepositExporter) FetchDepositData() error {
	// 1. 최근에 처리한 트랜잭션을 가져온다.
	db := db.GetConnection()
	var depositEvent *model.DepositEvent
	var lastSignature *solana.Signature

	err := db.Model(&model.DepositEvent{}).
		Order("slot DESC").
		First(&depositEvent).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		lastSignature = nil
	} else if err != nil {
		return errors.New("failed to get last depositEvent")
	} else {
		sig := solana.MustSignatureFromBase58(depositEvent.Signature)
		lastSignature = &sig
	}

	// 2. 해당 트랜잭션 이전의 모든 트랜잭션을 가져온다.
	rawTxSigs, err := e.getSignaturesForAddress(lastSignature)
	if err != nil {
		return errors.New("failed to get signatures for address")
	}

	// 3. Finalized 되지 않은 트랜잭션은 무시한다.
	txSigs := e.filterFinalizedSignatures(rawTxSigs)

	// 4. DepositEvent를 추출하고 account를 생성한다.
	for _, sig := range txSigs {
		tx, err := e.getTransaction(sig.Signature)
		if err != nil {
			return errors.New("failed to get transaction")
		}
		exist, address := e.extractDepositAddress(tx)
		if !exist {
			continue
		}

		if err := e.saveDepositAndCreateAccount(address, sig); err != nil {
			return errors.New("failed to save deposit and create account")
		}
	}

	return nil
}

func (e *DepositExporter) getSignaturesForAddress(lastSignature *solana.Signature) ([]*rpc.TransactionSignature, error) {
	ctx := context.Background()
	limit := 1000
	options := &rpc.GetSignaturesForAddressOpts{Limit: &limit}
	if lastSignature != nil {
		options.Until = *lastSignature
	}
	return e.client.GetSignaturesForAddressWithOpts(ctx, config.DEPOSIT_ADDRESS, options)
}

func (e *DepositExporter) getTransaction(txSig solana.Signature) (*rpc.GetTransactionResult, error) {
	ctx := context.Background()
	maxSupplortedTransactionVersion := uint64(2)
	return e.client.GetTransaction(ctx, txSig, &rpc.GetTransactionOpts{
		MaxSupportedTransactionVersion: &maxSupplortedTransactionVersion,
	})
}

func (e *DepositExporter) filterFinalizedSignatures(txSigs []*rpc.TransactionSignature) []*rpc.TransactionSignature {
	var finalizedSignatures []*rpc.TransactionSignature
	for _, sig := range txSigs {
		if sig.ConfirmationStatus == "finalized" {
			finalizedSignatures = append(finalizedSignatures, sig)
		}
	}
	return finalizedSignatures
}

// contract를 통해서 deposit 하는 경우는 무시한다.
func (e *DepositExporter) extractDepositAddress(tx *rpc.GetTransactionResult) (bool, string) {
	txLogs := tx.Meta.LogMessages
	logPattern := regexp.MustCompile(`Program log: Emitting DepositEvent: user=([^,]+),`)
	// txLogs 배열이 정확한 순서로 되어 있는지 확인하고, 조건이 맞으면 주소 추출
	for i := 0; i <= len(txLogs)-8; i++ {
		if txLogs[i] == "Program "+config.DEPOSIT_ADDRESS.String()+" invoke [1]" &&
			txLogs[i+1] == "Program log: Instruction: DepositSol" &&
			txLogs[i+2] == "Program 11111111111111111111111111111111 invoke [2]" &&
			txLogs[i+3] == "Program 11111111111111111111111111111111 success" &&
			strings.HasPrefix(txLogs[i+5], "Program data:") &&
			strings.HasPrefix(txLogs[i+6], "Program "+config.DEPOSIT_ADDRESS.String()+" consumed") &&
			txLogs[i+7] == "Program "+config.DEPOSIT_ADDRESS.String()+" success" {

			matches4 := logPattern.FindStringSubmatch(txLogs[i+4])

			if len(matches4) >= 1 {
				return true, matches4[1]
			}
		}
	}
	return false, ""
}

func (e *DepositExporter) saveDepositAndCreateAccount(address string, txSig *rpc.TransactionSignature) error {
	db := db.GetConnection()
	round := cache.GetRound()

	err := db.Transaction(func(tx *gorm.DB) error {
		// round와 address가 같고 deleted가 false인 account가 이미 있다면 deleted를 true로 업데이트
		if err := tx.Model(&model.Account{}).
			Where("round = ? AND address = ? AND deleted = ?", round, address, false).
			Update("deleted", true).Error; err != nil {
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
