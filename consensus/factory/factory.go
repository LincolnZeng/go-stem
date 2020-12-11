/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package factory

import (
	"crypto/ecdsa"
	"fmt"
	"path/filepath"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/common/errors"
	"github.com/scdoproject/go-stem/consensus"
	"github.com/scdoproject/go-stem/consensus/bft"
	"github.com/scdoproject/go-stem/consensus/bft/server"
	"github.com/scdoproject/go-stem/consensus/ethash"
	"github.com/scdoproject/go-stem/consensus/istanbul"
	"github.com/scdoproject/go-stem/consensus/istanbul/backend"
	"github.com/scdoproject/go-stem/consensus/pow"
	"github.com/scdoproject/go-stem/consensus/spow"
	"github.com/scdoproject/go-stem/database/leveldb"
)

// GetConsensusEngine get consensus engine according to miner algorithm name
// WARNING: engine may be a heavy instance. we should have as less as possible in our process.
func GetConsensusEngine(minerAlgorithm string, folder string) (consensus.Engine, error) {
	var minerEngine consensus.Engine
	if minerAlgorithm == common.EthashAlgorithm {
		minerEngine = ethash.New(ethash.GetDefaultConfig(), nil, false)
	} else if minerAlgorithm == common.Sha256Algorithm {
		minerEngine = pow.NewEngine(1)
	} else if minerAlgorithm == common.SpowAlgorithm {
		minerEngine = spow.NewSpowEngine(1, folder)
	} else {
		return nil, fmt.Errorf("unknown miner algorithm")
	}

	return minerEngine, nil
}

func GetBFTEngine(privateKey *ecdsa.PrivateKey, folder string) (consensus.Engine, error) {
	path := filepath.Join(folder, common.BFTDataFolder)
	db, err := leveldb.NewLevelDB(path)
	if err != nil {
		return nil, errors.NewStackedError(err, "create bft folder failed")
	}

	return backend.New(istanbul.DefaultConfig, privateKey, db), nil
}

func MustGetConsensusEngine(minerAlgorithm string) consensus.Engine {
	engine, err := GetConsensusEngine(minerAlgorithm, "temp")
	if err != nil {
		panic(err)
	}

	return engine
}

// subchain bft engine engine
// here need to input the privatekey
// TODO: not use privateKey as an input
func GetBFTSubchainEngine(privateKey *ecdsa.PrivateKey, folder string) (consensus.Engine, error) {
	path := filepath.Join(folder, common.BFTDataFolder)
	db, err := leveldb.NewLevelDB(path)
	if err != nil {
		return nil, errors.NewStackedError(err, "create bft folder failed")
	}

	return server.NewServer(bft.DefaultConfig, privateKey, db), nil
}
