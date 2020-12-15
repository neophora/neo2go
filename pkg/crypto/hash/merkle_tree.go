package hash

import (
	"errors"

	"github.com/neophora/neo2go/pkg/util"
)

// MerkleTree implementation.
type MerkleTree struct {
	root  *MerkleTreeNode
	depth int
}

// NewMerkleTree returns new MerkleTree object.
func NewMerkleTree(hashes []util.Uint256) (*MerkleTree, error) {
	if len(hashes) == 0 {
		return nil, errors.New("length of the hashes cannot be zero")
	}

	nodes := make([]*MerkleTreeNode, len(hashes))
	for i := 0; i < len(hashes); i++ {
		nodes[i] = &MerkleTreeNode{
			hash: hashes[i],
		}
	}

	return &MerkleTree{
		root:  buildMerkleTree(nodes),
		depth: 1,
	}, nil
}

// Root returns the computed root hash of the MerkleTree.
func (t *MerkleTree) Root() util.Uint256 {
	return t.root.hash
}

func buildMerkleTree(leaves []*MerkleTreeNode) *MerkleTreeNode {
	if len(leaves) == 0 {
		panic("length of leaves cannot be zero")
	}
	if len(leaves) == 1 {
		return leaves[0]
	}

	parents := make([]*MerkleTreeNode, (len(leaves)+1)/2)
	for i := 0; i < len(parents); i++ {
		parents[i] = &MerkleTreeNode{}
		parents[i].leftChild = leaves[i*2]
		leaves[i*2].parent = parents[i]

		if i*2+1 == len(leaves) {
			parents[i].rightChild = parents[i].leftChild
		} else {
			parents[i].rightChild = leaves[i*2+1]
			leaves[i*2+1].parent = parents[i]
		}

		b1 := parents[i].leftChild.hash.BytesBE()
		b2 := parents[i].rightChild.hash.BytesBE()
		b1 = append(b1, b2...)
		parents[i].hash = DoubleSha256(b1)
	}

	return buildMerkleTree(parents)
}

// MerkleTreeNode represents a node in the MerkleTree.
type MerkleTreeNode struct {
	hash       util.Uint256
	parent     *MerkleTreeNode
	leftChild  *MerkleTreeNode
	rightChild *MerkleTreeNode
}

// IsLeaf returns whether this node is a leaf node or not.
func (n *MerkleTreeNode) IsLeaf() bool {
	return n.leftChild == nil && n.rightChild == nil
}

// IsRoot returns whether this node is a root node or not.
func (n *MerkleTreeNode) IsRoot() bool {
	return n.parent == nil
}
