package server

import (
	"errors"
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/scdoproject/go-stem/common"
	"github.com/scdoproject/go-stem/consensus"
	"github.com/scdoproject/go-stem/consensus/bft"
	"github.com/scdoproject/go-stem/crypto"
	"github.com/scdoproject/go-stem/p2p"
)

/*this file will implement all methods at consensus/consensus.go Handler interface*/

const (
	bftMsg uint16 = 0x12
)

// define your errors here
var (
	errDecodeFailed = errors.New("fail to decode bft message")
)

// type handlerMessage struct {
// 	Code       uint16
// 	Size       uint32
// 	Payload    []byte
// 	ReceivedAt time.Time
// }

func (s *server) Protocal() consensus.Protocol {
	return consensus.Protocol{
		Name:     "bft",
		Versions: []uint{64}, //?
		Lengths:  []uint64{18},
	}
}

// // HandleMsg implements consensus.Handler.HandleMsg
func (s *server) HandleMsg(addr common.Address, message interface{}) (bool, error) {
	s.coreMu.Lock()
	defer s.coreMu.Unlock()

	msg, ok := message.(p2p.Message)
	if !ok {
		return false, errDecodeFailed
	}

	// make msg type is right
	if msg.Code == bftMsg {
		// if core is not started
		if !s.coreStarted {
			return true, bft.ErrEngineStopped
		}
		var data []byte
		if err := common.Deserialize(msg.Payload, &data); err != nil {
			s.log.Error("[DEBUG] consensus handler failed to deserialize msg")
			return true, errDecodeFailed
		}
		hash := crypto.HashBytes(data)

		// handle peer's message
		var m *lru.ARCCache
		ms, ok := s.recentMessages.Get(hash)

		if ok {
			m, _ = ms.(*lru.ARCCache)
		} else {
			m, _ = lru.NewARC(inmemoryMessages)
			s.recentMessages.Add(addr, m)
		}
		m.Add(hash, true)

		// handle self know message
		if _, ok := s.knownMessages.Get(hash); ok {
			return true, nil
		}
		s.knownMessages.Add(hash, true)

		go s.bftEventMux.Post(bft.MessageEvent{ // post all
			Payload: data,
		})
		fmt.Println("Post in HandleMsg")

		return true, nil
	}

	return false, nil
}

// HandleMsg implements consensus.Handler.HandleMsg
func (s *server) HandleMsg2(addr common.Address, msg p2p.Message) (bool, error) {
	s.coreMu.Lock()
	defer s.coreMu.Unlock()
	s.log.Error("[DEBUG-handleMsg] msg code %d", msg.Code)

	// make sure msg type is right
	if msg.Code == bftMsg {
		// if core is not started
		if !s.coreStarted {
			return true, bft.ErrEngineStopped
		}
		data, hash, err := s.decode(msg)
		if err != nil {
			return true, errDecodeFailed
		}

		// handle peer's message
		var m *lru.ARCCache
		ms, ok := s.recentMessages.Get(hash)

		if ok {
			m, _ = ms.(*lru.ARCCache)
		} else {
			m, _ = lru.NewARC(inmemoryMessages)
			s.recentMessages.Add(addr, m)
		}
		m.Add(hash, true)

		// handle self know message
		if _, ok := s.knownMessages.Get(hash); ok {
			return true, nil
		}
		s.knownMessages.Add(hash, true)

		go s.bftEventMux.Post(bft.MessageEvent{ // post all
			Payload: data,
		})
		fmt.Println("Post in HandleMsg")

		return true, nil
	}

	return false, nil
}

func (s *server) SetBroadcaster(broadcaster consensus.Broadcaster) {
	s.broadcaster = broadcaster
}

func (s *server) HandleNewChainHead() error {
	s.coreMu.RLock()
	defer s.coreMu.RUnlock()

	if !s.coreStarted {
		return bft.ErrEngineStopped
	}

	go s.bftEventMux.Post(bft.FinalCommittedEvent{})
	return nil
}

func (s *server) decode(msg p2p.Message) ([]byte, common.Hash, error) {
	var data []byte
	if err := msg.Decode(&data); err != nil {
		return nil, common.Hash{}, errDecodeFailed
	}
	return data, bft.RLPHash(data), nil
}
