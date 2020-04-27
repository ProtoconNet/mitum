package tree

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base/valuehash"
	jsonencoder "github.com/spikeekips/mitum/util/encoder/json"
)

type AVLTreeJSONPacker struct {
	jsonencoder.HintedHead
	TY string         `json:"tree_type"`
	RT string         `json:"root_key"`
	RH valuehash.Hash `json:"root_hash"`
	NS []Node         `json:"nodes"`
}

func (at *AVLTree) MarshalJSON() ([]byte, error) {
	if at == nil || at.Empty() {
		return jsonencoder.Marshal(nil)
	}

	var nodes []Node
	if err := at.Traverse(func(node Node) (bool, error) {
		nodes = append(nodes, node.Immutable())

		return true, nil
	}); err != nil {
		return nil, err
	}

	var rh valuehash.Hash
	if h, err := at.RootHash(); err != nil {
		return nil, err
	} else {
		rh = h
	}

	return jsonencoder.Marshal(AVLTreeJSONPacker{
		HintedHead: jsonencoder.NewHintedHead(at.Hint()),
		TY:         "avl hashable tree",
		RT:         string(at.Root().Key()),
		RH:         rh,
		NS:         nodes,
	})
}

type AVLTreeJSONUnpacker struct {
	RT string            `json:"root_key"`
	NS []json.RawMessage `json:"nodes"`
}

func (at *AVLTree) UnpackJSON(b []byte, enc *jsonencoder.Encoder) error {
	var uat AVLTreeJSONUnpacker
	if err := enc.Unmarshal(b, &uat); err != nil {
		return err
	}

	ns := make([][]byte, len(uat.NS))
	for i, b := range uat.NS {
		ns[i] = b
	}

	return at.unpack(enc, uat.RT, ns)
}
