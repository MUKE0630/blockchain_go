package main
 
//区块链保持一个有序快。
type BlockChain struct {
	blocks []*Block
}

//保存数据到链上
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	NewBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, NewBlock)
}

//使用创世纪块创建新的区块链
func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{NewGenesisBlock()}}
}
