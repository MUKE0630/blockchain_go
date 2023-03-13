# 数据结构

```go
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}
```



# 函数

```go
unc (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))		//找到blocks bucket
		encodedBlock := b.Get(i.currentHash)		//获取最后一个区块
		block = DeserializeBlock(encodedBlock)		//反序列化区块

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash		//将前一个区块的hash用于下一次寻找

	return block
}
```

