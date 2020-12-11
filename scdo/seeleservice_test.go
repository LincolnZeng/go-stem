/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package scdo

import (
	//"context"
	"context"
	"path/filepath"
	"testing"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus/factory"
	"github.com/scdoproject/go-stem/core"
	"github.com/scdoproject/go-stem/crypto"
	"github.com/scdoproject/go-stem/log"
	"github.com/scdoproject/go-stem/node"
	"github.com/stretchr/testify/assert"
)

func getTmpConfig() *node.Config {
	acctAddr := crypto.MustGenerateRandomAddress()

	return &node.Config{
		ScdoConfig: node.ScdoConfig{
			TxConf:   *core.DefaultTxPoolConfig(),
			Coinbase: *acctAddr,
		},
	}
}

func newTestScdoService() *ScdoService {
	conf := getTmpConfig()
	serviceContext := ServiceContext{
		DataDir: filepath.Join(common.GetTempFolder(), "n1"),
	}

	var key interface{} = "ServiceContext"
	ctx := context.WithValue(context.Background(), key, serviceContext)
	log := log.GetLogger("scdo")

	scdoService, err := NewScdoService(ctx, conf, log, factory.MustGetConsensusEngine(common.Sha256Algorithm), nil, -1)
	if err != nil {
		panic(err)
	}

	return scdoService
}

func Test_ScdoService_Protocols(t *testing.T) {
	s := newTestScdoService()
	defer s.Stop()

	protos := s.Protocols()
	assert.Equal(t, len(protos), 1)
}

func Test_ScdoService_Start(t *testing.T) {
	s := newTestScdoService()
	defer s.Stop()

	s.Start(nil)
	s.Stop()
	assert.Equal(t, s.scdoProtocol == nil, true)
}

func Test_ScdoService_Stop(t *testing.T) {
	s := newTestScdoService()
	defer s.Stop()

	s.Stop()
	assert.Equal(t, s.chainDB, nil)
	assert.Equal(t, s.accountStateDB, nil)
	assert.Equal(t, s.scdoProtocol == nil, true)

	// can be called more than once
	s.Stop()
	assert.Equal(t, s.chainDB, nil)
	assert.Equal(t, s.accountStateDB, nil)
	assert.Equal(t, s.scdoProtocol == nil, true)
}

func Test_ScdoService_APIs(t *testing.T) {
	s := newTestScdoService()
	apis := s.APIs()

	assert.Equal(t, len(apis), 10)
	assert.Equal(t, apis[0].Namespace, "scdo")
	assert.Equal(t, apis[1].Namespace, "txpool")
	assert.Equal(t, apis[2].Namespace, "network")
	assert.Equal(t, apis[3].Namespace, "debug")
	assert.Equal(t, apis[4].Namespace, "scdo")
	assert.Equal(t, apis[5].Namespace, "download")
	assert.Equal(t, apis[6].Namespace, "debug")
	assert.Equal(t, apis[7].Namespace, "miner")
	assert.Equal(t, apis[8].Namespace, "txpool")
}
