# 数据结构

```go
// 当前默克尔树
type MerkleTree struct {
	RootNode *MerkleNode
}

// 当前默克尔树节点
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}
```



# 函数

```go
// 从当前数据序列中创建新的默克尔树
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {		//如果交易数量是奇数
		data = append(data, data[len(data)-1])		//交易中复制最后一笔交易，凑偶数
	}

	for _, datum := range data {		//按序取出每笔交易
		node := NewMerkleNode(nil, nil, datum)		//对每笔交易创建单笔交易的默克尔节点
		nodes = append(nodes, *node)		//加入单笔交易节点集合
	}

	for i := 0; i < len(data)/2; i++ {		//构建默克尔树
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)		//以此取出“单笔交易节点集合”的前两个节点，构成倒数第二层节点
			newLevel = append(newLevel, *node)		//将倒数第二层节点加入新的节点集合
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}

// 创建一个新的默克尔树jie'dian
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.Data = hash[:]
	}

	mNode.Left = left
	mNode.Right = right

	return &mNode
}
```

