/*
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package bft

import (
	"github.com/scdoproject/go-stem/common"
)

type ProposerPolicy uint64

const (
	RoundRobin ProposerPolicy = iota // in a round robin setting, proposer will change in very block and round change.
	Sticky                           // with sticky property, propose will change only when a round change happens.
)

type BFTConfig struct {
	RequestTimeout uint64         `toml:",omitempty"` // The timeout for each Bft round in milliseconds.
	BlockPeriod    uint64         `toml:",omitempty"` // Default minimum difference between two consecutive block's timestamps in second
	ProposerPolicy ProposerPolicy `toml:",omitempty"` // The policy for proposer selection
	Epoch          uint64         `toml:",omitempty"` // The number of blocks after which to checkpoint and reset the pending votes
}

var DefaultConfig = &BFTConfig{
	RequestTimeout: 10000,                   // milliseconds
	BlockPeriod:    common.BFTBlockInterval, //second
	ProposerPolicy: RoundRobin,              // we use RoundRobin policy (others are Random/RoundRobin/LesatBusy/StickySession/cookies/ByReqeust)
	Epoch:          common.CheckInterval,    //blocks 1024
	// Epoch:          common.RelayRange,       //blocks 84*1024

}
