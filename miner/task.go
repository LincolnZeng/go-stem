/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package miner

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/common/memory"
	"github.com/scdoproject/go-stem/consensus"
	"github.com/scdoproject/go-stem/core"
	"github.com/scdoproject/go-stem/core/state"
	"github.com/scdoproject/go-stem/core/txs"
	"github.com/scdoproject/go-stem/core/types"
	"github.com/scdoproject/go-stem/database"
	"github.com/scdoproject/go-stem/log"
)

// Task is a mining work for engine, containing block header, transactions, and transaction receipts.
type Task struct {
	header   *types.BlockHeader
	txs      []*types.Transaction
	receipts []*types.Receipt
	debts    []*types.Debt

	coinbase     common.Address
	debtVerifier types.DebtVerifier
	// verifierTxs  []*types.Transaction
	// exitTxs      []*types.Transaction
	challengedTxs []*types.Transaction
	depositVers   []common.Address
	exitVers      []common.Address
}

// NewTask return Task object
func NewTask(header *types.BlockHeader, coinbase common.Address, verifier types.DebtVerifier) *Task {
	return &Task{
		header:       header,
		coinbase:     coinbase,
		debtVerifier: verifier,
	}
}

// applyTransactionsAndDebts TODO need to check more about the transactions, such as gas limit
func (task *Task) applyTransactionsAndDebts(scdo ScdoBackend, statedb *state.Statedb, accountStateDB database.Database, log *log.ScdoLog) error {
	now := time.Now()
	// entrance
	memory.Print(log, "task applyTransactionsAndDebts entrance", now, false)

	// choose transactions from the given txs
	var size int
	if task.header.Consensus != types.BftConsensus { // subchain doese not support debts.
		size = task.chooseDebts(scdo, statedb, log)
	}

	// the reward tx will always be at the first of the block's transactions
	reward, err := task.handleMinerRewardTx(statedb)
	if err != nil {
		return err
	}

	task.chooseTransactions(scdo, statedb, log, size)

	log.Info("mining block height:%d, reward:%s, transaction number:%d, debt number: %d",
		task.header.Height, reward, len(task.txs), len(task.debts))

	batch := accountStateDB.NewBatch()
	root, err := statedb.Commit(batch)
	if err != nil {
		return err
	}

	task.header.StateHash = root
	// task.header.SecondWitness =

	// exit
	memory.Print(log, "task applyTransactionsAndDebts exit", now, true)

	return nil
}

//applyTransactionsSubchain apply txs for subchain
func (task *Task) applyTransactionsSubchain(scdo ScdoBackend, statedb *state.Statedb, accountStateDB database.Database, log *log.ScdoLog, revertedTxHash *common.Hash) error {
	now := time.Now()
	// entrance
	memory.Print(log, "task applyTransactionsAndDebts entrance", now, false)

	// choose transactions from the given txs

	size := core.BlockByteLimit
	// the reward tx will always be at the first of the block's transactions
	reward, err := task.handleMinerRewardTx(statedb)
	if err != nil {
		return err
	}

	task.chooseTransactionsSubchain(scdo, statedb, log, size, revertedTxHash)

	log.Info("mining block height:%d, reward:%s, transaction number:%d, debt number: %d",
		task.header.Height, reward, len(task.txs), len(task.debts))

	batch := accountStateDB.NewBatch()
	root, err := statedb.Commit(batch)
	if err != nil {
		return err
	}

	task.header.StateHash = root

	// exit
	memory.Print(log, "task applyTransactionsAndDebts exit", now, true)

	return nil
}

func (task *Task) chooseDebts(scdo ScdoBackend, statedb *state.Statedb, log *log.ScdoLog) int {
	now := time.Now()
	// entrance
	memory.Print(log, "task chooseDebts entrance", now, false)

	size := core.BlockByteLimit

	for size > 0 {
		debts, _ := scdo.DebtPool().GetProcessableDebts(size)
		if len(debts) == 0 {
			break
		}

		for _, d := range debts {
			err := scdo.BlockChain().ApplyDebtWithoutVerify(statedb, d, task.coinbase)
			if err != nil {
				log.Warn("apply debt error %s", err)
				scdo.DebtPool().RemoveDebtByHash(d.Hash)
				continue
			}

			size = size - d.Size()
			task.debts = append(task.debts, d)
		}
	}

	// exit
	memory.Print(log, "task chooseDebts exit", now, true)

	return size
}

// handleMinerRewardTx handles the miner reward transaction.
func (task *Task) handleMinerRewardTx(statedb *state.Statedb) (*big.Int, error) {
	reward := consensus.GetReward(task.header.Height)
	rewardTx, err := txs.NewRewardTx(task.coinbase, reward, task.header.CreateTimestamp.Uint64())
	if err != nil {
		return nil, err
	}

	rewardTxReceipt, err := txs.ApplyRewardTx(rewardTx, statedb)
	if err != nil {
		return nil, err
	}

	task.txs = append(task.txs, rewardTx)

	// add the receipt of the reward tx
	task.receipts = append(task.receipts, rewardTxReceipt)

	return reward, nil
}

func (task *Task) chooseTransactions(scdo ScdoBackend, statedb *state.Statedb, log *log.ScdoLog, size int) {
	now := time.Now()
	// entrance
	memory.Print(log, "task chooseTransactions entrance", now, false)

	// TEST the event listner and fire function!
	// curHeight := task.header.Height
	// if curHeight%50 == 0 {
	// 	event.ChallengedTxEventManager.Fire(event.ChallengedTxEvent)
	// }

	//this code section for test the verifier is correctly added into secondwitness

	// task.depositVers = append(task.depositVers, common.BytesToAddress(hexutil.MustHexToBytes("0x1b9412d61a25f5f5decbf489fe5ed595d8b610a1")))
	// task.exitVers = append(task.exitVers, common.BytesToAddress(hexutil.MustHexToBytes("0x1b9412d61a25f5f5decbf489fe5ed595d8b610a1")))

	if len(task.depositVers) > 0 || len(task.exitVers) > 0 {
		log.Warn("deposit verifiers", task.depositVers)
		log.Warn("exit verifiers", task.exitVers)
		var err error
		task.header.SecondWitness, err = task.prepareWitness(task.header, task.challengedTxs, task.depositVers, task.exitVers)
		if err != nil {
			log.Error("failed to prepare deposit or exit tx into secondwitness")
		}
		log.Debug("apply new verifiers into witness, %s", task.header.SecondWitness)

	}
	// test code end here

	txIndex := 1 // the first tx is miner reward

	for size > 0 {
		txs, txsSize := scdo.TxPool().GetProcessableTransactions(size)
		if len(txs) == 0 {
			break
		}

		for _, tx := range txs {
			if err := tx.Validate(statedb, task.header.Height); err != nil {
				scdo.TxPool().RemoveTransaction(tx.Hash)
				log.Error("failed to validate tx %s, for %s", tx.Hash.Hex(), err)
				txsSize = txsSize - tx.Size()
				continue
			}

			receipt, err := scdo.BlockChain().ApplyTransaction(tx, txIndex, task.coinbase, statedb, task.header)
			if err != nil {
				scdo.TxPool().RemoveTransaction(tx.Hash)
				log.Error("failed to apply tx %s, %s", tx.Hash.Hex(), err)
				txsSize = txsSize - tx.Size()
				continue
			}
			if task.header.Consensus == types.BftConsensus { // for bft, the secondwitness will be used as deposit&exit address holder.
				rootAccounts := scdo.GenesisInfo().Rootaccounts
				fmt.Printf("rootAccounts %+v", rootAccounts)
				// if there is any successful challenge tx, need to revert blockchain first to specific point!
				if tx.IsChallengedTx(rootAccounts) {
					// will revert the block and db here, so the
					task.challengedTxs = append(task.challengedTxs, tx)
					// event.ChallengedTxEventManager.Fire(event.ChallengedTxEvent)
					return
				}

				if tx.IsVerifierTx(rootAccounts) {
					task.depositVers = append(task.depositVers, tx.ToAccount())
					// task.depositVers = append(task.depositVers, tx.FromAccount())
				}

				if tx.IsExitTx(rootAccounts) {
					task.exitVers = append(task.exitVers, tx.ToAccount())
				}
			}

			task.txs = append(task.txs, tx)
			task.receipts = append(task.receipts, receipt)
			txIndex++
		}
		size -= txsSize
	}
	if task.header.Consensus == types.BftConsensus {
		log.Info("[%d]deposit verifiers, [%d]exit verifiers, [%d]challenge txs", len(task.depositVers), len(task.exitVers), len(task.challengedTxs))
		var err error
		task.header.SecondWitness, err = task.prepareWitness(task.header, task.challengedTxs, task.depositVers, task.exitVers)
		if err != nil {
			log.Error("failed to prepare deposit or exit tx into secondwitness")
		}
		log.Debug("apply new verifiers into witness, %+v", task.header.SecondWitness)
	}

	// exit
	memory.Print(log, "task chooseTransactions exit", now, true)
}

func (task *Task) chooseTransactionsSubchain(scdo ScdoBackend, statedb *state.Statedb, log *log.ScdoLog, size int, revertedTxHash *common.Hash) {
	now := time.Now()
	// entrance
	memory.Print(log, "task chooseTransactions entrance", now, false)

	// TEST the event listner and fire function!
	// curHeight := task.header.Height
	// if curHeight%50 == 0 {
	// 	event.ChallengedTxEventManager.Fire(event.ChallengedTxEvent)
	// }

	//this code section for test the verifier is correctly added into secondwitness

	// task.depositVers = append(task.depositVers, common.BytesToAddress(hexutil.MustHexToBytes("0x1b9412d61a25f5f5decbf489fe5ed595d8b610a1")))
	// task.exitVers = append(task.exitVers, common.BytesToAddress(hexutil.MustHexToBytes("0x1b9412d61a25f5f5decbf489fe5ed595d8b610a1")))

	/*
		if len(task.depositVers) > 0 || len(task.exitVers) > 0 {
			log.Warn("deposit verifiers", task.depositVers)
			log.Warn("exit verifiers", task.exitVers)
			var err error
			task.header.SecondWitness, err = task.prepareWitness(task.header, task.challengedTxs, task.depositVers, task.exitVers)
			if err != nil {
				log.Error("failed to prepare deposit or exit tx into secondwitness")
			}
			log.Debug("apply new verifiers into witness, %s", task.header.SecondWitness)

		}
		// test code end here
	*/
	txIndex := 1 // the first tx is miner reward

	for size > 0 {
		txs, txsSize := scdo.TxPool().GetProcessableTransactions(size)
		if len(txs) == 0 {
			break
		}

		for _, tx := range txs {
			// expired := tx.largestPackHeight < task.header.Height
			// if expired {
			// 	fmt.Errorf("transaction must be packed at height %d but height %d, retreat!", tx.largestPackHeight, task.header.Height)
			// }
			if err := tx.Validate(statedb, task.header.Height); err != nil {
				scdo.TxPool().RemoveTransaction(tx.Hash)
				log.Error("failed to validate tx %s, for %s", tx.Hash.Hex(), err)
				txsSize = txsSize - tx.Size()
				continue
			}

			receipt, err := scdo.BlockChain().ApplyTransaction(tx, txIndex, task.coinbase, statedb, task.header)
			if err != nil {
				scdo.TxPool().RemoveTransaction(tx.Hash)
				log.Error("failed to apply tx %s, %s", tx.Hash.Hex(), err)
				txsSize = txsSize - tx.Size()
				continue
			}
			if task.header.Consensus == types.BftConsensus { // for bft, the secondwitness will be used as deposit&exit address holder.
				rootAccounts := scdo.GenesisInfo().Rootaccounts
				fmt.Printf("rootAccounts %+v", rootAccounts)
				// if there is any successful challenge tx, need to revert blockchain first to specific point!
				if tx.IsChallengedTx(rootAccounts) {
					// will revert the block and db here, so the
					task.challengedTxs = append(task.challengedTxs, tx)
					if !tx.Hash.Equal(*revertedTxHash) { // this tx challenge is not handled before, fire the event to handle it
						*revertedTxHash = tx.Hash
						// event.ChallengedTxEventManager.Fire(event.ChallengedTxEvent)
						panic("Alert: subchain is challenged successfully, stop!")
						return
					}
				}

				if tx.IsVerifierTx(rootAccounts) {
					task.depositVers = append(task.depositVers, tx.ToAccount())
					// task.depositVers = append(task.depositVers, tx.FromAccount())
				}
				// test code
				// task.depositVers = append(task.depositVers, common.BytesToAddress(hexutil.MustHexToBytes("0x42cab76173e399824422ae12c63f6f8c84e6ca71")))
				// end test code
				if tx.IsExitTx(rootAccounts) {
					task.exitVers = append(task.exitVers, tx.ToAccount())
				}
			}

			task.txs = append(task.txs, tx)
			task.receipts = append(task.receipts, receipt)
			txIndex++
		}
		size -= txsSize
	}
	if task.header.Consensus == types.BftConsensus {
		log.Info("[%d]deposit verifiers, [%d]exit verifiers, [%d]challenge txs", len(task.depositVers), len(task.exitVers), len(task.challengedTxs))
		var err error
		task.header.SecondWitness, err = task.prepareWitness(task.header, task.challengedTxs, task.depositVers, task.exitVers)
		if err != nil {
			log.Error("failed to prepare deposit or exit tx into secondwitness")
		}
		log.Debug("apply new verifiers into witness, %+v", task.header.SecondWitness)
	}

	// exit
	memory.Print(log, "task chooseTransactions exit", now, true)
}

// generateBlock builds a block from task
func (task *Task) generateBlock() *types.Block {
	return types.NewBlock(task.header, task.txs, task.receipts, task.debts)
}

// Result is the result mined by engine. It contains the raw task and mined block.
type Result struct {
	task  *Task
	block *types.Block // mined block, with good nonce
}

// prepareWitness prepare header witness for deposit(header.Witness) or exit(header.SecondWitness)
// func (task *Task) prepareWitness(header *types.BlockHeader, depositVers []common.Address, exitVers []common.Address) ([]byte, error) {
func (task *Task) prepareWitness(header *types.BlockHeader, chTxs []*types.Transaction, depositVers []common.Address, exitVers []common.Address) ([]byte, error) {
	var buf bytes.Buffer
	// compensate the lack bytes if header.Extra is not enough BftExtraVanity bytes.

	if len(header.SecondWitness) < types.BftExtraVanity { //here we use BftExtraVanity (32-bit fixed length)
		header.SecondWitness = append(header.SecondWitness, bytes.Repeat([]byte{0x00}, types.BftExtraVanity-len(header.SecondWitness))...)
	}
	buf.Write(header.SecondWitness[:types.BftExtraVanity])

	updatedVers := &types.SecondWitnessExtra{ // we share the BftExtra struct
		ChallengedTxs: chTxs,
		DepositVers:   depositVers,
		ExitVers:      exitVers,
	}

	payload, err := rlp.EncodeToBytes(&updatedVers)
	if err != nil {
		return nil, err
	}
	return append(buf.Bytes(), payload...), nil
}
