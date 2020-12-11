/**
*  @file
*  @copyright defined in go-seele/LICENSE
 */

package core

import (
	"github.com/scdoproject/go-stem/common"
)

func (c *core) handleFinalCommitted() error {
	c.logger.Debug("Received a final committed proposal")
	c.startNewRound(common.Big0)
	return nil
}
