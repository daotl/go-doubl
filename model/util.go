package model

import (
	"bytes"
	"errors"
	"unsafe"

	"github.com/crpt/go-crpt"
	cbg "github.com/daotl/cbor-gen"
	"github.com/daotl/go-marsha"
)

var (
	ErrInvalidBytes = errors.New("failed to unmarshal because of invalid bytes")
)

// Util provides utility methods to work with DOUBL models.
type Util struct {
	Mrsh marsha.Marsha
	Crpt crpt.Crpt
}

// New creates a new Util with the specified Marsha and Crpt instances.
func New(mrsh marsha.Marsha, crpt crpt.Crpt) *Util {
	return &Util{
		Mrsh: mrsh,
		Crpt: crpt,
	}
}

// TransactionsFromExtPtrs returns the Transactions wrapped in the TransactionExtSlice.
func TransactionSliceFromExtSlice(txxs TransactionExtSlice) TransactionSlice {
	txs := make(TransactionSlice, len(txxs))
	for i, bhx := range txxs {
		txs[i] = *bhx.Transaction
	}
	return txs
}

// BlockHeadersFromExts returns the BlockHeaders wrapped in the BlockHeaderExts.
func BlockHeadersFromExts(bhxs []BlockHeaderExt) []BlockHeader {
	bhs := make([]BlockHeader, len(bhxs))
	for i, bhx := range bhxs {
		bhs[i] = *bhx.BlockHeader
	}
	return bhs
}

// BlocksFromExts returns the Blocks wrapped in the BlockExts.
func BlocksFromExts(bxs []BlockExt) []Block {
	bs := make([]Block, len(bxs))
	for i, bhx := range bxs {
		bs[i] = *bhx.Raw()
	}
	return bs
}

// ExtendTransaction extends a Transaction into a TransactionExt.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) ExtendTransaction(tx *Transaction) (*TransactionExt, error) {
	bin, err := u.Mrsh.MarshalStruct(tx)
	if err != nil {
		return nil, err
	}
	return &TransactionExt{
		Transaction: tx,
		Bytes:       bin,
		Hash:        u.Crpt.Hash(bin),
	}, nil
}

// ExtendTransaction extends TransactionSlice into TransactionExtSlice.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) ExtendTransactionSlice(txs TransactionSlice) (TransactionExtSlice, error) {
	txxs := make(TransactionExtSlice, len(txs))
	var err error
	for i := range txs {
		if txxs[i], err = u.ExtendTransaction(&txs[i]); err != nil {
			return nil, err
		}
	}
	return txxs, nil
}

// TransactionExtFromBytes unmarshal Transaction and extend it into a TransactionExt.
// TransactionExt.Bytes points to the same underlying memory as `bin` for performance consideration,
// it's not safe to modify it anywhere.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) TransactionExtFromBytes(bin []byte) (*TransactionExt, error) {
	txx := &TransactionExt{
		Transaction: new(Transaction),
	}
	read, err := u.Mrsh.UnmarshalStruct(bin, txx.Transaction)
	if err != nil {
		return nil, err
	}
	txx.Bytes = bin[:read]
	txx.Hash = u.Crpt.Hash(bin[:read])
	return txx, err
}

// TransactionExtSliceFromBytesSlice unmarshal Transaction byte slices and extend them into TransactionExtSlice.
// TransactionExt.Bytes points to the same underlying memory as `bins` for performance consideration,
// it's not safe to modify it anywhere.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) TransactionExtSliceFromBytesSlice(bins [][]byte) (TransactionExtSlice, error) {
	txxs := make(TransactionExtSlice, len(bins))
	var err error
	for i, bin := range bins {
		if txxs[i], err = u.TransactionExtFromBytes(bin); err != nil {
			return nil, err
		}
	}
	return txxs, nil
}

// TransactionExtSliceFromTransactionsBytes unmarshal a single Transactions bytes into
// TransactionSlice and extend them into TransactionExtSlice. maxCount should be equal or larger
// than the transaction count for the best performance, but it's not necessary.
// TransactionExt.Bytes points to the same underlying memory as `bins` for performance consideration,
// it's not safe to modify it anywhere.
func (u *Util) TransactionExtSliceFromTransactionsBytesWithMaxCount(bin []byte, maxCount int,
) (TransactionExtSlice, error) {
	txxs := make(TransactionExtSlice, 0, maxCount)
	for i := 0; i < len(bin); {
		txx, err := u.TransactionExtFromBytes(bin[i:])
		if err != nil {
			return nil, err
		}
		txxs = append(txxs, txx)
		i += len(txx.Bytes)
	}
	return txxs, nil
}

// TransactionExtSliceFromTransactionsBytes unmarshal a single Transactions bytes into
// TransactionSlice and extend them into TransactionExtSlice.
// TransactionExt.Bytes points to the same underlying memory as `bins` for performance consideration,
// it's not safe to modify it anywhere.
func (u *Util) TransactionExtSliceFromTransactionsBytes(bin []byte) (TransactionExtSlice, error) {
	return u.TransactionExtSliceFromTransactionsBytesWithMaxCount(bin, 0)
}

// HashTransaction computes the hash of the transaction.
func (u *Util) HashTransaction(tx *Transaction) (TransactionHash, error) {
	bin, err := u.Mrsh.MarshalStruct(tx)
	if err != nil {
		return nil, err
	}
	return u.Crpt.Hash(bin), nil
}

// HashTransactionNoSig computes the hash of the transaction without signature.
func (u *Util) HashTransactionNoSig(tx *Transaction) (TransactionHash, error) {
	txNoSig := getTxNoSig(tx)
	bin, err := u.Mrsh.MarshalStruct(txNoSig)
	if err != nil {
		return nil, err
	}
	return u.Crpt.Hash(bin), nil
}

// HashTransactionSlice computes the hashes of the transactions.
func (u *Util) HashTransactionSlice(txs TransactionSlice) (hs []TransactionHash, err error) {
	hs = make([]TransactionHash, len(txs))
	for i, val := range txs {
		if hs[i], err = u.HashTransaction(&val); err != nil {
			return nil, err
		}
	}
	return hs, nil
}

// VerifyTransactionSignature verifies the transaction signature.
// Should prefer using VerifyTransactionExtSignature instead for better performance.
// TODO: prepend genesis block hash to bin
func (u *Util) VerifyTransactionSignature(tx *Transaction) (bool, error) {
	sig := tx.Sig
	//fmt.Printf("sig: %0x\n", sig)
	txNoSig := getTxNoSig(tx)
	bin, err := u.Mrsh.MarshalStruct(txNoSig)
	if err != nil {
		return false, err
	}
	//fmt.Printf("bin: %0x\n", bin)

	pub, err := u.Crpt.PublicKeyFromBytes(tx.From)
	if err != nil {
		return false, err
	}
	//fmt.Printf("pub: %0x\n", pub)
	return pub.VerifyMessage(bin, sig)
}

// 0x41=64
const signatureCborDataLengthByte = byte(SignatureCborDataLength)

// VerifyTransactionExtSignature verifies the transaction signature from TransactionExt.
// TODO: Prepend genesis block hash to bin
func (u *Util) VerifyTransactionExtSignature(txx *TransactionExt) (bool, error) {
	// In CBOR-encoded Transaction bytes:
	// if the signature is set, it's encoded as a byte string with txNoSigLen 64;
	// if not, it's a byte string with txNoSigLen 0, not `null`
	txNoSigLen := len(txx.Bytes) - SignatureCborDataLength - 1
	txNoSigBytes := make([]byte, txNoSigLen)
	copy(txNoSigBytes, txx.Bytes[:txNoSigLen-1])
	txNoSigBytes[txNoSigLen-1] = signatureCborDataLengthByte

	pub, err := u.Crpt.PublicKeyFromBytes(txx.From)
	if err != nil {
		return false, err
	}
	return pub.VerifyMessage(txNoSigBytes, txx.Sig)
}

// HashBlockHeader computes the hash of the BlockHeader.
func (u *Util) HashBlockHeader(bh *BlockHeader) (BlockHash, error) {
	bin, err := u.Mrsh.MarshalStruct(bh)
	if err != nil {
		return nil, err
	}
	return u.Crpt.Hash(bin), nil
}

// ExtendBlockHeader extends a BlockHeader into a BlockHeaderExt.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) ExtendBlockHeader(bh *BlockHeader) (*BlockHeaderExt, error) {
	bin, err := u.Mrsh.MarshalStruct(bh)
	if err != nil {
		return nil, err
	}
	return &BlockHeaderExt{
		BlockHeader: bh,
		Bytes:       bin,
		Hash:        u.Crpt.Hash(bin),
	}, nil
}

// BlockHeaderExtFromBytes unmarshal BlockHeader and extend it into a BlockHeaderExt.
// BlockHeaderExt.Bytes points to the same underlying memory as `bins` for performance consideration,
// it's not safe to modify it anywhere.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) BlockHeaderExtFromBytes(bin []byte) (*BlockHeaderExt, error) {
	bhx := &BlockHeaderExt{
		BlockHeader: new(BlockHeader),
	}
	read, err := u.Mrsh.UnmarshalStruct(bin, bhx.BlockHeader)
	if err != nil {
		return nil, err
	}
	bhx.Bytes = bin[:read]
	bhx.Hash = u.Crpt.Hash(bin[:read])
	return bhx, err
}

// ExtendBlock extends a Block into a BlockExt.
func (u *Util) ExtendBlock(b *Block) (*BlockExt, error) {
	bhx, err := u.ExtendBlockHeader(&b.Header)
	if err != nil {
		return nil, err
	}

	txxs, err := u.ExtendTransactionSlice(b.Txs)
	if err != nil {
		return nil, err
	}

	return &BlockExt{
		block:  b,
		Header: bhx,
		Txs:    txxs,
	}, nil
}

// ExtractBlockHeaderExtFromBlockBytes extract bytes corresponding to the BlockHeader from Block bytes
// and unmarshal it into a BlockHeaderExt.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) ExtractBlockHeaderExtFromBlockBytes(bin []byte) (*BlockHeaderExt, error) {
	return u.BlockHeaderExtFromBytes(bin[BlockCborInitialLength:])
}

// BlockExtFromBytes unmarshal Block and extend it into a BlockExt.
// BlockHeaderExt.Bytes and TransactionExt.Bytes point to the same underlying memory as `bins` for
// performance consideration, it's not safe to modify them anywhere.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) BlockExtFromBytes(bin []byte) (bx *BlockExt, err error) {
	bx = &BlockExt{}

	bx.Header, err = u.ExtractBlockHeaderExtFromBlockBytes(bin)
	if err != nil {
		return nil, err
	}

	i := BlockCborInitialLength + len(bx.Header.Bytes)
	maj, l, read, err := cbg.CborReadHeader(bytes.NewReader(bin[i:]))
	if err != nil || maj != cbg.MajArray {
		return nil, ErrInvalidBytes
	}

	if l > 0 {
		bx.Txs, err = u.TransactionExtSliceFromTransactionsBytesWithMaxCount(bin[i+read:], int(l))
		if err != nil {
			return nil, err
		}
	}

	return bx, nil
}

// GenRootHashFromTransactionHashes computes the merkle tree root hash from TransactionHashes.
func (u *Util) GenRootHashFromTransactionHashes(transactionHashes []TransactionHash) TransactionsRootHash {
	return u.Crpt.MerkleHashFromByteSlices(*(*[][]byte)(unsafe.Pointer(&transactionHashes)))
}

// GenRootHashFromTransactions computes the merkle tree root hash from TransactionSlice.
func (u *Util) GenRootHashFromTransactionSlice(txs TransactionSlice) (TransactionsRootHash, error) {
	txhs := make([][]byte, len(txs))
	var err error
	for i := range txs {
		if txhs[i], err = u.HashTransaction(&txs[i]); err != nil {
			return nil, err
		}
	}
	return u.Crpt.MerkleHashFromByteSlices(txhs), nil
}

// GenRootHashFromTransactionExts computes the merkle tree root hash from TransactionExtSlice.
func (u *Util) GenRootHashFromTransactionExts(txxs TransactionExtSlice) TransactionsRootHash {
	txhs := make([][]byte, len(txxs))
	for i, txx := range txxs {
		txhs[i] = txx.Hash
	}
	return u.Crpt.MerkleHashFromByteSlices(txhs)
}

// If tx contains signature, return a copy without the signature.
func getTxNoSig(tx *Transaction) (txNoSig *Transaction) {
	// Don't need to copy Transaction
	if tx.Sig == nil {
		return tx
	}

	txNoSig = new(Transaction)
	*txNoSig = *tx
	txNoSig.Sig = nil
	return txNoSig
}
