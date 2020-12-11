package core

import "github.com/scdoproject/go-stem/common"

func (c *core) handleFinalCommit() error {
	c.log.Debug("received a final commit proposal")
	c.startNewRound(common.Big0)
	return nil
}
