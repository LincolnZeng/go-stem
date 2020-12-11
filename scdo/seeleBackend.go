package scdo

import (
	"math/big"

	"github.com/scdoproject/go-stem/api"
	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/core/store"
	"github.com/scdoproject/go-stem/core/types"
	"github.com/scdoproject/go-stem/log"
	"github.com/scdoproject/go-stem/p2p"
	"github.com/scdoproject/go-stem/scdo/download"
)

type SeeleBackend struct {
	s *SeeleService
}

// NewSeeleBackend backend
func NewSeeleBackend(s *SeeleService) *SeeleBackend {
	return &SeeleBackend{s}
}

// TxPoolBackend tx pool
func (sd *SeeleBackend) TxPoolBackend() api.Pool { return sd.s.txPool }

// GetNetVersion net version
func (sd *SeeleBackend) GetNetVersion() string { return sd.s.netVersion }

// GetNetWorkID net id
func (sd *SeeleBackend) GetNetWorkID() string { return sd.s.networkID }

// GetP2pServer p2p server
func (sd *SeeleBackend) GetP2pServer() *p2p.Server { return sd.s.p2pServer }

// ChainBackend block chain db
func (sd *SeeleBackend) ChainBackend() api.Chain { return sd.s.chain }

// Log return log pointer
func (sd *SeeleBackend) Log() *log.SeeleLog { return sd.s.log }

// IsSyncing check status
func (sd *SeeleBackend) IsSyncing() bool {
	scdoserviceAPI := sd.s.APIs()[5]
	d := scdoserviceAPI.Service.(downloader.PrivatedownloaderAPI)

	return d.IsSyncing()
}

// ProtocolBackend return protocol
func (sd *SeeleBackend) ProtocolBackend() api.Protocol { return sd.s.scdoProtocol }

// GetBlock returns the requested block by hash or height
func (sd *SeeleBackend) GetBlock(hash common.Hash, height int64) (*types.Block, error) {
	var block *types.Block
	var err error
	if !hash.IsEmpty() {
		store := sd.s.chain.GetStore()
		block, err = store.GetBlock(hash)
		if err != nil {
			return nil, err
		}
	} else {
		if height < 0 {
			header := sd.s.chain.CurrentHeader()
			block, err = sd.s.chain.GetStore().GetBlockByHeight(header.Height)
		} else {
			block, err = sd.s.chain.GetStore().GetBlockByHeight(uint64(height))
		}
		if err != nil {
			return nil, err
		}
	}

	return block, nil
}

// GetBlockTotalDifficulty return total difficulty
func (sd *SeeleBackend) GetBlockTotalDifficulty(hash common.Hash) (*big.Int, error) {
	store := sd.s.chain.GetStore()
	return store.GetBlockTotalDifficulty(hash)
}

// GetReceiptByTxHash get receipt by transaction hash
func (sd *SeeleBackend) GetReceiptByTxHash(hash common.Hash) (*types.Receipt, error) {
	store := sd.s.chain.GetStore()
	receipt, err := store.GetReceiptByTxHash(hash)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

// GetTransaction return tx
func (sd *SeeleBackend) GetTransaction(pool api.PoolCore, bcStore store.BlockchainStore, txHash common.Hash) (*types.Transaction, *api.BlockIndex, error) {
	return api.GetTransaction(pool, bcStore, txHash)
}
