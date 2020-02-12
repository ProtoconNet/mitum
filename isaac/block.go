package isaac

import (
	"time"

	"github.com/spikeekips/mitum/hint"
	"github.com/spikeekips/mitum/isvalid"
	"github.com/spikeekips/mitum/valuehash"
)

var (
	BlockType          hint.Type = hint.Type([2]byte{0x05, 0x00})
	BlockOperationType hint.Type = hint.Type([2]byte{0x05, 0x02})
	BlockStatesType    hint.Type = hint.Type([2]byte{0x05, 0x03})
	BlockStateType     hint.Type = hint.Type([2]byte{0x05, 0x04})
)

type Block interface {
	isvalid.IsValider
	hint.Hinter
	Bytes() []byte
	Hash() valuehash.Hash // root hash
	PreviousBlock() valuehash.Hash
	Height() Height
	Round() Round
	Proposal() valuehash.Hash
	Operations() valuehash.Hash
	States() valuehash.Hash
	INITVoteProof() VoteProof
	ACCEPTVoteProof() VoteProof
	CreatedAt() time.Time
}

type BlockOperations interface {
	isvalid.IsValider
	hint.Hinter
	Bytes() []byte
	Hash() valuehash.Hash
	Operations() []Operation
}

type BlockStates interface {
	isvalid.IsValider
	hint.Hinter
	Bytes() []byte
	Hash() valuehash.Hash
	States() []BlockState
}

type BlockState interface {
	isvalid.IsValider
	hint.Hinter
	Bytes() []byte
	Hash() valuehash.Hash
	Value() interface{} // TODO BlockStateValue interface{}
	PreviousBlock() valuehash.Hash
}