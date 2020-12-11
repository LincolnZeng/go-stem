/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package core

import (
	"github.com/scdoproject/go-stem/consensus/istanbul"
)

type backlogEvent struct {
	src istanbul.Validator
	msg *message
}

type timeoutEvent struct{}
