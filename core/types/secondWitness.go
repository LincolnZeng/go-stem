package types

import (
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/scdoproject/go-stem/common"
)

type SecondWitnessExtra struct {
	ChallengedTxs []*Transaction
	DepositVers   []common.Address
	ExitVers      []common.Address
}

func (swExtra *SecondWitnessExtra) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, []interface{}{
		swExtra.DepositVers,
		swExtra.ExitVers,
	})
}

func (swExtra *SecondWitnessExtra) DecodeRLP(s *rlp.Stream) error {
	var secondWitnessExtra struct {
		DepositVers []common.Address
		ExitVers    []common.Address
	}
	if err := s.Decode(&secondWitnessExtra); err != nil {
		return err
	}
	swExtra.DepositVers, swExtra.ExitVers = secondWitnessExtra.DepositVers, secondWitnessExtra.ExitVers
	return nil
}

// ExtractSWExtra extract verifiers from SecondWitness
func (swExtra *SecondWitnessExtra) ExtractSWExtra(h *BlockHeader) error {
	// if len(h.ExtraData) < BftExtraVanity {
	// 	fmt.Printf("header extra data len %d is smaller than BftExtraVanity %d\n", len(h.ExtraData), BftExtraVanity)
	// 	return nil, ErrInvalidBftHeaderExtra
	// }

	// var swExtra SecondWitnessExtra
	err := rlp.DecodeBytes(h.SecondWitness[:], &swExtra)
	if err != nil {
		fmt.Println("DecodeBytes err, ", err)
		return err
	}
	return nil
}

func ExtractSWExtra(h *BlockHeader) (*SecondWitnessExtra, error) {
	// if len(h.ExtraData) < BftExtraVanity {
	// 	fmt.Printf("header extra data len %d is smaller than BftExtraVanity %d\n", len(h.ExtraData), BftExtraVanity)
	// 	return nil, ErrInvalidBftHeaderExtra
	// }
	fmt.Println("decode swextra", h.SecondWitness)
	var bftExtra *SecondWitnessExtra
	err := rlp.DecodeBytes(h.ExtraData, &bftExtra)
	if err != nil {
		return nil, err
	}
	return bftExtra, nil
}

// func ExtractSWExtra(h *BlockHeader, val interface{}) error {
// 	return rlp.DecodeBytes(h.SecondWitness[:], val)
// }

func ExtractSecondWitnessExtra(h *BlockHeader) (*SecondWitnessExtra, error) {
	if h.Height == 0 {
		return nil, nil
	}
	if len(h.ExtraData) < BftExtraVanity {
		return nil, ErrInvalidBftHeaderExtra
	}
	var swExtra *SecondWitnessExtra
	err := rlp.DecodeBytes(h.SecondWitness[BftExtraVanity:], &swExtra)
	if err != nil {
		return nil, err
	}
	return swExtra, nil
}
