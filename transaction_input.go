package main

import(
	"bytes"
)

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey []byte
}

//检查地址初始化交易
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
