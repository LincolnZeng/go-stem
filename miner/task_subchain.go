/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package miner

// import (
// 	"math/big"
// 	"time"

// 	"github.com/scdoproject/go-stem/common"
// 	"github.com/scdoproject/go-stem/common/memory"
// 	"github.com/scdoproject/go-stem/consensus"
// 	"github.com/scdoproject/go-stem/core"
// 	"github.com/scdoproject/go-stem/core/state"
// 	"github.com/scdoproject/go-stem/core/txs"
// 	"github.com/scdoproject/go-stem/core/types"
// 	"github.com/scdoproject/go-stem/database"
// 	"github.com/scdoproject/go-stem/log"
// )

// // Task is a mining work for engine, containing block header, transactions, and transaction receipts.
// type TaskSubChain struct {
// 	header           *types.BlockHeader
// 	txs              []*types.Transaction
// 	receipts         []*types.Receipt
// 	updatedVerifiers []common.Address
// 	coinbase         common.Address
// }

// // NewTask return TaskSubChain object
// func NewTaskSubChain(header *types.BlockHeader, coinbase common.Address, vers []common.Address) *TaskSubChain {
// 	return &TaskSubChain{
// 		header:           header,
// 		coinbase:         coinbase,
// 		updatedVerifiers: vers,
// 	}
// }

// // applyTransactionsAndDebts TODO need to check more about the transactions, such as gas limit
// func (task *TaskSubChain) applyTransactions(seele SeeleBackend, statedb *state.Statedb, accountStateDB database.Database, log *log.SeeleLog) error {
// 	now := time.Now()
// 	// entrance
// 	memory.Print(log, "task applyTransactionsAndDebts entrance", now, false)

// 	// choose transactions from the given txs
// 	size := core.BlockByteLimit
// 	// the reward tx will always be at the first of the block's transactions
// 	reward, err := task.handleMinerRewardTx(statedb)
// 	if err != nil {
// 		return err
// 	}

// 	task.chooseTransactions(seele, statedb, log, size)

// 	log.Info("mining block height:%d, reward:%s, transaction number:%d, updated verifiers number: %d",
// 		task.header.Height, reward, len(task.txs), len(task.updatedVerifiers))

// 	batch := accountStateDB.NewBatch()
// 	root, err := statedb.Commit(batch)
// 	if err != nil {
// 		return err
// 	}

// 	task.header.StateHash = root

// 	// exit
// 	memory.Print(log, "task applyTransactionsAndDebts exit", now, true)

// 	return nil
// }

// func (task *TaskSubChain) chooseDebts(seele SeeleBackend, statedb *state.Statedb, log *log.SeeleLog) int {
// 	now := time.Now()
// 	// entrance
// 	memory.Print(log, "task chooseDebts entrance", now, false)

// 	size := core.BlockByteLimit

// 	for size > 0 {
// 		debts, _ := seele.DebtPool().GetProcessableDebts(size)
// 		if len(debts) == 0 {
// 			break
// 		}

// 		for _, d := range debts {
// 			err := seele.BlockChain().ApplyDebtWithoutVerify(statedb, d, task.coinbase)
// 			if err != nil {
// 				log.Warn("apply debt error %s", err)
// 				seele.DebtPool().RemoveDebtByHash(d.Hash)
// 				continue
// 			}

// 			size = size - d.Size()
// 			task.debts = append(task.debts, d)
// 		}
// 	}

// 	// exit
// 	memory.Print(log, "task chooseDebts exit", now, true)

// 	return size
// }

// // handleMinerRewardTx handles the miner reward transaction.
// func (task *TaskSubChain) handleMinerRewardTx(statedb *state.Statedb) (*big.Int, error) {
// 	reward := consensus.GetReward(task.header.Height)
// 	rewardTx, err := txs.NewRewardTx(task.coinbase, reward, task.header.CreateTimestamp.Uint64())
// 	if err != nil {
// 		return nil, err
// 	}

// 	rewardTxReceipt, err := txs.ApplyRewardTx(rewardTx, statedb)
// 	if err != nil {
// 		return nil, err
// 	}

// 	task.txs = append(task.txs, rewardTx)

// 	// add the receipt of the reward tx
// 	task.receipts = append(task.receipts, rewardTxReceipt)

// 	return reward, nil
// }

// func (task *TaskSubChain) chooseTransactions(seele SeeleBackend, statedb *state.Statedb, log *log.SeeleLog, size int) {
// 	now := time.Now()
// 	// entrance
// 	memory.Print(log, "task chooseTransactions entrance", now, false)

// 	txIndex := 1 // the first tx is miner reward

// 	for size > 0 {
// 		txs, txsSize := seele.TxPool().GetProcessableTransactions(size)
// 		if len(txs) == 0 {
// 			break
// 		}

// 		for _, tx := range txs {
// 			if err := tx.Validate(statedb, task.header.Height); err != nil {
// 				seele.TxPool().RemoveTransaction(tx.Hash)
// 				log.Error("failed to validate tx %s, for %s", tx.Hash.Hex(), err)
// 				txsSize = txsSize - tx.Size()
// 				continue
// 			}

// 			receipt, err := seele.BlockChain().ApplyTransaction(tx, txIndex, task.coinbase, statedb, task.header)
// 			if err != nil {
// 				seele.TxPool().RemoveTransaction(tx.Hash)
// 				log.Error("failed to apply tx %s, %s", tx.Hash.Hex(), err)
// 				txsSize = txsSize - tx.Size()
// 				continue
// 			}

// 			task.txs = append(task.txs, tx)
// 			task.receipts = append(task.receipts, receipt)
// 			txIndex++
// 		}

// 		size -= txsSize
// 	}

// 	// exit
// 	memory.Print(log, "task chooseTransactions exit", now, true)
// }

// // generateBlock builds a block from task
// func (task *TaskSubChain) generateBlock() *types.Block {
// 	return types.NewBlock(task.header, task.txs, task.receipts, task.debts)
// }

// // Result is the result mined by engine. It contains the raw task and mined block.
// type Result struct {
// 	task  *TaskSubChain
// 	block *types.Block // mined block, with good nonce
// }
