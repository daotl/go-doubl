package util

import (
	"unsafe"

	"github.com/crpt/go-crpt"
	"github.com/daotl/go-marsha"

	. "github.com/daotl/go-doubl/model"
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

// ExtendTransaction extends Transactions into []*TransactionExt.
// NOTE: ExtraUnmarshaled is not set yet.
func (u *Util) ExtendTransactions(txs Transactions) ([]*TransactionExt, error) {
	txxs := make([]*TransactionExt, len(txs))
	var err error
	for i := range txs {
		if txxs[i], err = u.ExtendTransaction(&txs[i]); err != nil {
			return nil, err
		}
	}
	return txxs, nil
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

// HashTransactions computes the hashes of the transactions.
func (u *Util) HashTransactions(txs []Transaction) (hs []TransactionHash, err error) {
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

// ExtendBlock extends a Block into a BlockExt.
func (u *Util) ExtendBlock(b *Block) (*BlockExt, error) {
	bhx, err := u.ExtendBlockHeader(&b.Header)
	if err != nil {
		return nil, err
	}

	txxs, err := u.ExtendTransactions(b.Txs)
	if err != nil {
		return nil, err
	}

	return &BlockExt{
		Block:  b,
		Header: bhx,
		Txs:    txxs,
	}, nil
}

// GenRootHashFromTransactionHashes computes the merkle tree root hash from TransactionHashes.
func (u *Util) GenRootHashFromTransactionHashes(transactionHashes []TransactionHash) TransactionsRootHash {
	return u.Crpt.MerkleHashFromByteSlices(*(*[][]byte)(unsafe.Pointer(&transactionHashes)))
}

// GenRootHashFromTransactions computes the merkle tree root hash from Transactions.
func (u *Util) GenRootHashFromTransactions(txs []Transaction) (TransactionsRootHash, error) {
	txhs := make([][]byte, len(txs))
	var err error
	for i := range txs {
		if txhs[i], err = u.HashTransaction(&txs[i]); err != nil {
			return nil, err
		}
	}
	return u.Crpt.MerkleHashFromByteSlices(txhs), nil
}

// GenRootHashFromTransactionExts computes the merkle tree root hash from []*TransactionExt.
func (u *Util) GenRootHashFromTransactionExts(txxs []*TransactionExt) TransactionsRootHash {
	txHashes := make([][]byte, len(txxs))
	for i, txx := range txxs {
		txHashes[i] = txx.Hash
	}
	return u.Crpt.MerkleHashFromByteSlices(txHashes)
}

// If tx contains signature, return a copy without the signature.
func getTxNoSig(tx *Transaction) (txNoSig *Transaction) {
	// Don't need to copy Transaction
	if tx.Sig != nil {
		return tx
	}

	txNoSig = new(Transaction)
	*txNoSig = *tx
	txNoSig.Sig = nil
	return txNoSig
}
