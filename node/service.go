/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package node

import (
	"github.com/scdoproject/go-stem/p2p"
	"github.com/scdoproject/go-stem/rpc"
)

// Service represents a service which is registered to the node after the node starts.
type Service interface {
	// Protocols retrieves the P2P protocols the service wishes to start.
	Protocols() []p2p.Protocol

	APIs() (apis []rpc.API)

	Start(server *p2p.Server) error

	Stop() error
}
