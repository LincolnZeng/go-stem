/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package core

import (
	"math/big"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus/pow"
	"github.com/scdoproject/go-stem/core/store"
	"github.com/scdoproject/go-stem/core/types"
	"github.com/scdoproject/go-stem/database/leveldb"
)

func newTestGenesis() *Genesis {
	accounts := map[common.Address]*big.Int{
		types.TestGenesisAccount.Addr: types.TestGenesisAccount.Amount,
	}

	return GetGenesis(NewGenesisInfo(accounts, 1, 0, big.NewInt(0), types.PowConsensus, nil))
}

func NewTestBlockchain() *Blockchain {
	return NewTestBlockchainWithVerifier(nil)
}

func NewTestBlockchainWithVerifier(verifier types.DebtVerifier) *Blockchain {
	db, _ := leveldb.NewTestDatabase()

	bcStore := store.NewCachedStore(store.NewBlockchainDatabase(db))

	genesis := newTestGenesis()
	if err := genesis.InitializeAndValidate(bcStore, db); err != nil {
		panic(err)
	}

	bc, err := NewBlockchain(bcStore, db, "", pow.NewEngine(1), verifier, -1)
	if err != nil {
		panic(err)
	}

	return bc
}
