/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package monitor

import (
	"github.com/scdoproject/go-stem/log"
	"github.com/scdoproject/go-stem/node"
	"github.com/scdoproject/go-stem/p2p"
	rpc "github.com/scdoproject/go-stem/rpc"
	"github.com/scdoproject/go-stem/scdo"
)

// MonitorService implements some rpc interfaces provided by a monitor server
type MonitorService struct {
	p2pServer *p2p.Server        // Peer-to-Peer server infos
	scdo      *scdo.SeeleService // scdo full node service
	seeleNode *node.Node         // scdo node
	log       *log.SeeleLog

	rpcAddr string // listening port
	name    string // name displayed on the moitor
	node    string // node name
	version string // version
}

// NewMonitorService returns a MonitorService instance
func NewMonitorService(seeleService *scdo.SeeleService, seeleNode *node.Node, conf *node.Config, slog *log.SeeleLog, name string) (*MonitorService, error) {
	return &MonitorService{
		scdo:      seeleService,
		seeleNode: seeleNode,
		log:       slog,
		name:      name,
		rpcAddr:   conf.BasicConfig.RPCAddr,
		node:      conf.BasicConfig.Name,
		version:   conf.BasicConfig.Version,
	}, nil
}

// Protocols implements node.Service, return nil as it dosn't use the p2p service
func (s *MonitorService) Protocols() []p2p.Protocol { return nil }

// Start implements node.Service, starting goroutines needed by SeeleService.
func (s *MonitorService) Start(srvr *p2p.Server) error {
	s.p2pServer = srvr

	s.log.Info("monitor rpc service started")

	return nil
}

// Stop implements node.Service, terminating all internal goroutines.
func (s *MonitorService) Stop() error {

	return nil
}

// APIs implements node.Service, returning the collection of RPC services the scdo package offers.
func (s *MonitorService) APIs() (apis []rpc.API) {
	return append(apis, []rpc.API{
		{
			Namespace: "monitor",
			Version:   "1.0",
			Service:   NewPublicMonitorAPI(s),
			Public:    true,
		},
	}...)
}
