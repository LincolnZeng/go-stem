/**
*  @file
*  @copyright defined in go-stem/LICENSE
 */

package light

import (
	"testing"

	"github.com/scdoproject/go-stem/common"
	"github.com/stretchr/testify/assert"
)

func Test_odrTriePoof_Rlp(t *testing.T) {
	proof := odrTriePoof{
		Root:  common.StringToHash("root"),
		Key:   []byte("trie key"),
		Proof: make([]proofNode, 0),
	}

	encoded, err := common.Serialize(proof)
	assert.Nil(t, err)

	proof2 := odrTriePoof{}
	err = common.Deserialize(encoded, &proof2)
	assert.Nil(t, err)
	assert.Equal(t, proof, proof2)
}
