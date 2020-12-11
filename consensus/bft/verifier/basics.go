package verifier

import (
	"reflect"
	"sort"
	"sync"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus/bft"
)

//////////////////////////////////////////////////////
// basicVerifier
type basicVerifier struct {
	address common.Address
}

func (ver *basicVerifier) Address() common.Address {
	return ver.address
}

func (ver *basicVerifier) String() string {
	return ver.Address().String()
}

/////////////////////////////////////////////////////
// basicSet
type basicSet struct {
	verifiers  bft.Verifiers
	policy     bft.ProposerPolicy
	proposer   bft.Verifier
	verifierMu sync.RWMutex
	selector   bft.ProposalSelector
}

// newBasicSet create verSet with addresses, policy, selector
func newBasicSet(addrs []common.Address, policy bft.ProposerPolicy) *basicSet {
	verSet := &basicSet{}
	verSet.policy = policy
	// init verifiers
	verSet.verifiers = make([]bft.Verifier, len(addrs))
	for i, addr := range addrs {
		verSet.verifiers[i] = NewVerifier(addr)
		// fmt.Printf("[VerSet] add add %+v into verset %+v", addr, verSet.verifiers[i])
	}
	//sort
	sort.Sort(verSet.verifiers)
	// set the first verifier as the proposer
	if verSet.Size() > 0 {
		verSet.proposer = verSet.GetVerByIndex(0)
	}
	verSet.selector = roundRobinProposer // we use roound robin policy to select proposer
	if policy == bft.Sticky {
		verSet.selector = stickyProposer
	}
	return verSet
}

// Size return the size of basicSet
func (verSet *basicSet) Size() int {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	return len(verSet.verifiers)
}

// GetVerByIndex get verifier by the index
func (verSet *basicSet) GetVerByIndex(i uint64) bft.Verifier {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	if i < uint64(verSet.Size()) {
		return verSet.verifiers[i]
	}
	return nil
}

// GetVerByAddress get verifier by address
func (verSet *basicSet) GetVerByAddress(addr common.Address) (int, bft.Verifier) {
	// for i, ver := range verSet.List() {
	// 	if addr == ver.Address() {
	// 		return i, ver
	// 	}
	// }

	// for j, veri := range verSet.verifiers {
	// 	fmt.Printf("[%d]th verifier [%s]\n", j, veri.Address())
	// }
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	for i, ver := range verSet.verifiers {
		if addr == ver.Address() {
			return i, ver
		}
	}
	return -1, nil
}

// proposer methods
func (verSet *basicSet) GetProposer() bft.Verifier {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	return verSet.proposer
}

func (verSet *basicSet) IsProposer(address common.Address) bool {
	_, val := verSet.GetVerByAddress(address)
	return reflect.DeepEqual(verSet.GetProposer(), val)
}

////// proposer-related policy   ///////
func roundRobinProposer(verSet bft.VerifierSet, proposer common.Address, round uint64) bft.Verifier {
	if verSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if isEmptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(verSet, proposer, round) + 1
	}
	pick := seed % uint64(verSet.Size())
	return verSet.GetVerByIndex(pick)
}

func stickyProposer(verSet bft.VerifierSet, proposer common.Address, round uint64) bft.Verifier {
	if verSet.Size() == 0 {
		return nil
	}
	seed := uint64(0)
	if isEmptyAddress(proposer) {
		seed = round
	} else {
		seed = calcSeed(verSet, proposer, round)
	}
	pick := seed % uint64(verSet.Size())
	return verSet.GetVerByIndex(pick)
}

func (verSet *basicSet) Policy() bft.ProposerPolicy {
	return verSet.policy
}

func (verSet *basicSet) CalcProposer(lastProposer common.Address, round uint64) {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	verSet.proposer = verSet.selector(verSet, lastProposer, round)
}

func calcSeed(verSet bft.VerifierSet, proposer common.Address, round uint64) uint64 {
	offset := 0
	if idx, val := verSet.GetVerByAddress(proposer); val != nil {
		offset = idx
	}
	return uint64(offset) + round
}

func isEmptyAddress(addr common.Address) bool {
	return addr == common.Address{}
}

func (verSet *basicSet) List() []bft.Verifier {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()
	return verSet.verifiers
}

// verifier-related methods
func (verSet *basicSet) AddVerifier(address common.Address) bool {
	verSet.verifierMu.Lock()
	defer verSet.verifierMu.Unlock()
	// check existence
	for _, v := range verSet.verifiers {
		if v.Address() == address { // verifier already in the list, just return false
			return false
		}
	}
	verSet.verifiers = append(verSet.verifiers, NewVerifier(address))
	sort.Sort(verSet.verifiers)
	return true
}

// RemoveVerifier remove address from verifiers
func (verSet *basicSet) RemoveVerifier(address common.Address) bool {
	verSet.verifierMu.Lock()
	defer verSet.verifierMu.Unlock()

	// fmt.Println("To remove", address, "from verifiers set", verSet.verifiers)

	for i, v := range verSet.verifiers {
		if v.Address() == address {
			verSet.verifiers = append(verSet.verifiers[:i], verSet.verifiers[i+1:]...)
			return true
		}
	}
	return false
}

func (verSet *basicSet) Copy() bft.VerifierSet {
	verSet.verifierMu.RLock()
	defer verSet.verifierMu.RUnlock()

	addresses := make([]common.Address, 0, len(verSet.verifiers))
	for _, v := range verSet.verifiers {
		addresses = append(addresses, v.Address())
	}
	return NewVerifierSet(addresses, verSet.policy)
}

// failure tolerate
func (verSet *basicSet) F() int {
	return int((float64(verSet.Size()) / 3)) - 1 // 1 stands for self
}
