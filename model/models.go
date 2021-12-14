package model

import (
	"unsafe"

	"github.com/daotl/go-marsha"
	"github.com/daotl/go-marsha/cborgen"
)

const (
	// Reference: https://datatracker.ietf.org/doc/html/rfc7049#section-2

	CborNull       = byte(0xf6)
	CborNullLength = 1

	// Transaction is serialized as a CBOR array
	// Transaction initial byte: major type 4 (100) + map length 6 (00110)
	TransactionCborInitial = byte(0b100_00110)

	// Signature initial 2 bytes:
	// 0x58: 0b010_11000: major type 2 (010) + additional information 24 means 1-byte length (11000)
	// 0x41: 1-byte length: 64 bytes
	SignatureCborInitialLength = 2
	SignatureCborDataLength    = 64
)

// Transaction defines a transaction that can be encoded into CBOR or JSON representation.
type Transaction struct {

	// Transaction type
	Type TransactionType `json:"type"`

	// Sender address
	From Address `json:"from"`

	// Nonce
	Nonce uint64 `json:"nonce"`

	// Target address (optional)
	To Address `json:"to,omitempty"`

	// Payload data (optional)
	Data []byte `json:"data,omitempty"`

	// Extra metadata (optional)
	// This can be used for e.g., encryption with secret sharing scheme
	Extra []byte `json:"extra,omitempty"`

	// Signature (omitted when calculating the signature)
	// Put it last in the CBOR array, so it can be efficiently appended.
	Sig Signature `json:"signature,omitempty"`
}

// Ptr implements marsha.Struct
func (t Transaction) Ptr() marsha.StructPtr { return &t }

// Val implements marsha.StructPtr
func (t *Transaction) Val() marsha.Struct { return *t }

// Size is the estimated occupied memory of Transaction.
func (t *Transaction) Size() int {
	return int(unsafe.Sizeof(t)) + len(t.From) + len(t.To) + len(t.Data) + len(t.Sig)
}

// Transactions is a slice of Transaction
type Transactions []Transaction

// Val implements marsha.StructSlicePtr
func (ms *Transactions) Val() []marsha.StructPtr {
	models := make([]marsha.StructPtr, 0, len(*ms))
	for i := range *ms {
		models = append(models, &(*ms)[i])
	}
	return models
}

// NewStructPtr implements cborgen.StructSlicePtr
func (*Transactions) NewStructPtr() marsha.StructPtr { return new(Transaction) }

// Append implements cborgen.StructSlicePtr
func (ms *Transactions) Append(m cborgen.StructPtr) { *ms = append(*ms, *(m.(*Transaction))) }

// TransactionExt extends Transaction to hold some useful data that can be computed once and used in multiple places.
type TransactionExt struct {
	*Transaction

	// CBOR encoded Transaction
	Bytes []byte

	// Transaction hash
	Hash TransactionHash

	// Pointer to the unmarshaled Transaction.Extra field
	ExtraUnmarshaled *interface{}
}

// Size is the estimated occupied memory of TransactionExt.
func (t *TransactionExt) Size() int {
	return int(unsafe.Sizeof(t)) + t.Transaction.Size() + len(t.Bytes) + len(t.Hash)
}

// BlockHeader defines a block header that can be encoded into CBOR or JSON representation.
type BlockHeader struct {

	// Creator
	Creator Address `json:"creator"`

	// Timestamp in Unix time
	Time Timestamp `json:"timestamp"`

	// Previous block hashes
	PrevHashes []BlockHash `json:"prevHashes,omitempty"`

	// Block height
	Height BlockHeight `json:"height"`

	// Root hash of some kind of hash-based tree (e.g., Merkle tree) of transactions
	TxRoot TransactionsRootHash `json:"transactionsRoot"`

	// Number of transactions contained in the block
	TxCount uint64 `json:"transactionCount"`

	// Extra metadata (optional)
	// This can be used for e.g., consensus-specific data
	Extra []byte `json:"extra,omitempty"`

	// Signature (omitted when calculating the signature)
	// Put it last in the CBOR array, so it can be efficiently appended.
	Sig Signature `json:"signature,omitempty"`
}

// Ptr implements marsha.Struct
func (bh BlockHeader) Ptr() marsha.StructPtr { return &bh }

// Val implements marsha.StructPtr
func (bh *BlockHeader) Val() marsha.Struct { return *bh }

// BlockHeaderExt extends BlockHeader to hold some useful data that can be computed once and used in multiple places.
type BlockHeaderExt struct {
	*BlockHeader

	// CBOR encoded BlockHeader
	Bytes []byte

	// Block hash (the same as BlockHeader hash)
	Hash BlockHash

	// Pointer to the unmarshaled Transaction.Extra field
	ExtraUnmarshaled *interface{}
}

// Block defines a block that can be encoded into CBOR or JSON representation.
type Block struct {

	// Block header
	Header BlockHeader `json:"header"`

	// Transactions contained in the block
	Txs Transactions `json:"transactions,omitempty"`
}

// Ptr implements marsha.Struct
func (b Block) Ptr() marsha.StructPtr { return &b }

// Val implements marsha.StructPtr
func (b *Block) Val() marsha.Struct { return *b }

// BlockExt extends Block to hold some useful data that can be computed once and used in multiple places.
type BlockExt struct {
	*Block

	// Extended block header
	Header *BlockHeaderExt

	// Extended transactions
	Txs []*TransactionExt
}
