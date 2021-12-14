//go:generate go run github.com/daotl/go-doubl/test/cborgen

package test

import (
	crand "crypto/rand"
	"math/rand"
	"strconv"
	"time"

	"github.com/crpt/go-crpt"
	"github.com/daotl/go-marsha"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"

	. "github.com/daotl/go-doubl/model"
	"github.com/daotl/go-doubl/util"
)

func init() {
	SetUtil(Util)
	GenTestModels()
}

type TransactionNoSig struct {
	Type  TransactionType
	From  Address
	Nonce uint64
	To    Address
	Data  []byte
	Extra []byte
}

func (t TransactionNoSig) Ptr() marsha.StructPtr { return &t }
func (p *TransactionNoSig) Val() marsha.Struct   { return *p }

var (
	ut *util.Util

	TestPublicKey       crpt.PublicKey
	TestPrivateKey      crpt.PrivateKey
	TestAddress         crpt.Address
	TestSignature       Signature
	TestTransaction     Transaction
	TestTransactionHash TransactionHash
	TestTransactions    Transactions
	TestBlockHeader     BlockHeader

	TestHash = Hash32{
		0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4,
		0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4, 0x1, 0x2, 0x3, 0x4,
	}
	TestHash2 = Hash32{
		0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8,
		0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8, 0x5, 0x6, 0x7, 0x8,
	}
	TestBlockHash   = TestHash
	TestBlockHashes = make([]BlockHash, 1)
	TestAddress2    = Address(TestHash)
)

func SetUtil(u *util.Util) {
	ut = u
}

// Gen test models
func GenTestModels() {
	TestPublicKey, TestPrivateKey, _ = ut.Crpt.GenerateKey(nil)
	TestAddress = TestPublicKey.Address()
	TestTransaction = Transaction{
		Type:  4,
		Nonce: 52,
		From:  TestAddress,
		To:    TestAddress2,
		Data:  []byte{0x4, 0x13, 0x52},
		Extra: []byte{0x21, 0x9, 0x09},
		Sig:   nil,
	}

	bin, err := ut.Mrsh.MarshalStruct(&TestTransaction)
	if err != nil {
		panic(err)
	}

	TestTransaction.Sig, err = TestPrivateKey.SignMessage(bin, nil)
	if err != nil {
		panic(err)
	}

	TestTransactionHash, err = ut.HashTransaction(&TestTransaction)
	if err != nil {
		panic(err)
	}

	TestTransactions = Transactions{TestTransaction, TestTransaction}

	TestBlockHeader = GenTestBlockHeaderWithExtra([]byte{0x4, 0x13, 0x52})
}

// GenRandomTransaction generates a random Transaction for test.
// NOTE: Sig is not a valid signature.
func GenRandomTransaction() *Transaction {
	return &Transaction{
		Type:  4,
		From:  TestAddress,
		Nonce: 52,
		To:    TestAddress2,
		Data:  []byte(strconv.FormatInt(rand.Int63(), 10)),
		Extra: []byte(strconv.FormatInt(rand.Int63(), 10)),
		Sig:   TestSignature,
	}
}

// GenRandomTransactionExt generates a random TransactionExt for test.
func GenRandomTransactionExt() *TransactionExt {
	tx := GenRandomTransaction()
	txx, err := ut.ExtendTransaction(tx)
	if err != nil {
		panic(err)
	}
	return txx
}

// GenRandomTransactionExt generates random TransactionExts for test.
func GenRandomTransactionExts(min, max int) []TransactionExt {
	min, max = adjustRange(min, max)
	n := min + rand.Int()%(max-min+1)
	txxs := make([]TransactionExt, n)
	for i := 0; i < n; i++ {
		tx := GenRandomTransactionExt()
		txxs = append(txxs, *tx)
	}
	return txxs
}

// GenTestBlockHeaderWithExtra creates a BlockHeader for test with given Extra data.
func GenTestBlockHeaderWithExtra(extraBytes []byte) BlockHeader {
	return BlockHeader{
		Creator:    TestAddress,
		Time:       1525392000,
		PrevHashes: []BlockHash{TestHash},
		Height:     52,
		TxRoot:     TestHash,
		TxCount:    52,
		Extra:      extraBytes,
		Sig:        TestSignature,
	}
}

// GenRandomHash generates a random hash for test.
func GenRandomHash() Hash32 {
	a := [HashSize]byte{}
	crand.Read(a[:])
	return a[:]
}

// GenRandomHash generates random hashes for test.
func GenRandomHashes() []Hash32 {
	rand.Seed(time.Now().UnixNano())
	cnt := rand.Intn(100)
	hs := make([]Hash32, cnt)
	for i := 0; i < cnt; i++ {
		hs[i] = GenRandomHash()
	}
	return hs
}

// GenRandomCid generates a random CID for test.
func GenRandomCid() cid.Cid {
	rh := GenRandomHash()
	mhbytes, _ := mh.Encode(rh, mh.SHA3_256)
	return cid.NewCidV1(cid.DagCBOR, mhbytes)
}

/* Helpers */

func adjustRange(min, max int) (int, int) {
	switch {
	case max < 0:
		max = 0
		fallthrough
	case min < 0:
		min = 0
		fallthrough
	case max < min:
		min = max
	}
	return min, max
}
