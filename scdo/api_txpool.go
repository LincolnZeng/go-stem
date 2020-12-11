/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package scdo

import (
	api2 "github.com/scdoproject/go-stem/api"
	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/common/hexutil"
)

// TransactionPoolAPI provides an API to access transaction pool information.
type TransactionPoolAPI struct {
	s *ScdoService
}

// NewTransactionPoolAPI creates a new PrivateTransactionPoolAPI object for transaction pool rpc service.
func NewTransactionPoolAPI(s *ScdoService) *TransactionPoolAPI {
	return &TransactionPoolAPI{s}
}

// GetDebtByHash return the debt info by debt hash
func (api *TransactionPoolAPI) GetDebtByHash(debtHash string) (map[string]interface{}, error) {
	hashByte, err := hexutil.HexToBytes(debtHash)
	if err != nil {
		return nil, err
	}
	hash := common.BytesToHash(hashByte)

	debt, blockIdx, err := api2.GetDebt(api.s.DebtPool(), api.s.chain.GetStore(), hash)
	if err != nil {
		return nil, err
	}

	output := map[string]interface{}{
		"debt": debt,
	}

	if blockIdx == nil {
		output["status"] = "pool"
	} else {
		output["status"] = "block"
		output["blockHash"] = blockIdx.BlockHash.Hex()
		output["blockHeight"] = blockIdx.BlockHeight
		output["debtIndex"] = blockIdx.Index
	}

	return output, nil
}
