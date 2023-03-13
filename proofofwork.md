# 数据结构

```go
//最大随机数
var maxNonce = math.MaxInt64
// 挖矿难度
const targetBits = 24
type ProofOfWork struct {
	block  *Block
	target *big.Int
}
```



# 函数

```go
// 创建并返回工作量证明
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))	//Lsh 将target左移

	pow := &ProofOfWork{b, target}		//创建工作量证明的数据结构pow

	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(		//Join 连接slice 成为一个新的 slice
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),		//IntToHex 将int64转为字节数组
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// 运行工作量证明
func (pow *ProofOfWork) Run() (int, []byte) {	
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining a new block")
	for nonce < maxNonce {		//调整随机值
		data := pow.prepareData(nonce)		//预处理

		hash = sha256.Sum256(data)		//计算sha256
		if math.Remainder(float64(nonce), 100000) == 0 {		//Reminder 返回 nonce/100000 的浮点值，该值等于零则
			fmt.Printf("\r%x", hash)
		}
		hashInt.SetBytes(hash[:])		//将字节数组转换为大端整形。

		if hashInt.Cmp(pow.target) == -1 {		//该hash值 < 目标值则成功
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// 
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

```

