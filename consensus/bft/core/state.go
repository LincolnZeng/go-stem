package core

import (
	"math/big"
	"sync"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus/bft"
)

type roundState struct {
	round          *big.Int
	sequence       *big.Int
	Preprepare     *bft.Preprepare
	Prepares       *messageSet
	Commits        *messageSet
	lockedHash     common.Hash
	pendingRequest *bft.Request

	mu             *sync.RWMutex
	hasBadProposal func(hash common.Hash) bool
}

// newRoundState creates a new roundState instance with the given view and VerifierSet
// lockedHash and preprepare are for round change when lock exists,
// we need to keep a reference of preprepare in order to propose locked proposal when there is a lock and itself is the proposer
func newRoundState(view *bft.View, verifierSet bft.VerifierSet, lockedHash common.Hash, preprepare *bft.Preprepare, pendingRequest *bft.Request, hasBadProposal func(hash common.Hash) bool) *roundState {
	return &roundState{
		round:          view.Round,
		sequence:       view.Sequence,
		Preprepare:     preprepare,
		Prepares:       newMessageSet(verifierSet),
		Commits:        newMessageSet(verifierSet),
		lockedHash:     lockedHash,
		mu:             new(sync.RWMutex),
		pendingRequest: pendingRequest,
		hasBadProposal: hasBadProposal,
	}
}

// LockHash lock the proposal hash for whole round.
// lockedHash in order to make sure proposal and commit are on the same round
func (s *roundState) LockHash() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Preprepare != nil {
		s.lockedHash = s.Preprepare.Proposal.Hash()
	}
}

func (s *roundState) UnlockHash() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lockedHash = common.Hash{}
}

func (s *roundState) Sequence() *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sequence
}

func (s *roundState) IsHashLocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if common.EmptyHash == s.lockedHash {
		return false
	}
	return !s.hasBadProposal(s.GetLockedHash())
}

func (s *roundState) GetLockedHash() common.Hash {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lockedHash
}

func (s *roundState) Round() *big.Int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.round
}

func (s *roundState) Subject() *bft.Subject {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Preprepare == nil {
		return nil
	}

	return &bft.Subject{
		View: &bft.View{
			Round:    new(big.Int).Set(s.round),
			Sequence: new(big.Int).Set(s.sequence),
		},
		Digest: s.Preprepare.Proposal.Hash(),
	}
}

func (s *roundState) Proposal() bft.Proposal {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Preprepare != nil {
		return s.Preprepare.Proposal
	}

	return nil
}

func (s *roundState) GetPrepareOrCommitSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := s.Prepares.Size() + s.Commits.Size()

	// find duplicate one
	for _, m := range s.Prepares.Values() {
		if s.Commits.Get(m.Address) != nil {
			result--
		}
	}
	return result
}

func (s *roundState) SetPreprepare(preprepare *bft.Preprepare) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Preprepare = preprepare
}
