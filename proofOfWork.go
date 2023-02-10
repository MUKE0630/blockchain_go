package main

import(
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var maxNonce=math.MaxInt64

//挖矿难度.
const targetBits=24

type ProofOfWork struct{
	block *Block
	target *big.Int
}

//构建新的工作量证明并返回
func NewProofOfWork(b*Block)*ProofOfWork{
	target:=big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))
	pow:=&ProofOfWork{b,target}
	return pow
}
	

func (pow *ProofOfWork) prepareData(nonce int) []byte{
	data:=bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
			
		},
		[]byte{},
	)
	return data
}

//执行工作量证明
func (pow *ProofOfWork) RunProofOfWork() (int,[]byte){
	var hashInt big.Int
	var hash [32]byte
	nonce:=0

	fmt.Print("Mining the block containing.....\"&s\"\n",pow.block.Data)
	for nonce<maxNonce{
		data:=pow.prepareData(nonce)
		hash=sha256.Sum256(data)
		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target)==-1{
			break
		}else{
			nonce++
		}
	}
	fmt.Print("\n\n")
	return nonce,hash[:]
}

//工作验证
func (pow *ProofOfWork) Validate() bool{
	var hashInt big.Int

	data:=pow.prepareData(pow.block.Nonce)
	hash:=sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid:=hashInt.Cmp(pow.target)==-1
	return isValid
}