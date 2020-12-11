package core

import (
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus/bft"
)

type messageSet struct {
	view       *bft.View
	verSet     bft.VerifierSet
	messagesMu *sync.Mutex
	messages   map[common.Address]*message
}

func newMessageSet(verSet bft.VerifierSet) *messageSet {
	return &messageSet{
		view: &bft.View{
			Round:    new(big.Int),
			Sequence: new(big.Int),
		},
		messagesMu: new(sync.Mutex),
		messages:   make(map[common.Address]*message),
		verSet:     verSet,
	}
}

func (ms *messageSet) Size() int {
	ms.messagesMu.Lock()
	defer ms.messagesMu.Unlock()
	return len(ms.messages)
}

func (ms *messageSet) Add(msg *message) error {
	ms.messagesMu.Lock()
	defer ms.messagesMu.Unlock()

	if err := ms.verify(msg); err != nil {
		return err
	}

	return ms.addVerifiedMessage(msg)
}

func (ms *messageSet) Values() (result []*message) {
	ms.messagesMu.Lock()
	defer ms.messagesMu.Unlock()

	for _, v := range ms.messages {
		result = append(result, v)
	}

	return result
}

func (ms *messageSet) verify(msg *message) error {
	if _, v := ms.verSet.GetVerByAddress(msg.Address); v == nil {
		return bft.ErrAddressUnauthorized
	}
	// TODO check view number and sequence number
	return nil
}

func (ms *messageSet) addVerifiedMessage(msg *message) error {
	ms.messages[msg.Address] = msg
	return nil
}

func (ms *messageSet) String() string {
	ms.messagesMu.Lock()
	defer ms.messagesMu.Unlock()
	addresses := make([]string, 0, len(ms.messages))
	for _, v := range ms.messages {
		addresses = append(addresses, v.Address.String())
	}
	return fmt.Sprintf("[%v]", strings.Join(addresses, ", "))
}

func (ms *messageSet) Get(address common.Address) *message {
	ms.messagesMu.Lock()
	defer ms.messagesMu.Unlock()
	return ms.messages[address]
}
