package model_test

import (
	"bytes"
	"fmt"
	"testing"

	ipldcbor "github.com/ipfs/go-ipld-cbor"
	"github.com/stretchr/testify/require"

	. "github.com/daotl/go-doubl/model"
	"github.com/daotl/go-doubl/test"
)

func TestMarshalUnmarshal(t *testing.T) {
	req := require.New(t)
	mrsh := test.Mrsh

	t.Run("Transaction.MarshalCBOR/UnmarshalCBOR", func(t *testing.T) {
		tx := &test.TestTransaction
		var b bytes.Buffer
		err := tx.MarshalCBOR(&b)
		req.NoError(err)
		bin := b.Bytes()
		fmt.Println("Serialized Transaction size: ", len(bin))
		fmt.Printf("%02x\n", bin)
		tx2 := &Transaction{}
		err = tx2.UnmarshalCBOR(&b)
		req.NoError(err)
		req.Equal(tx, tx2)
	})

	t.Run("Marshal & unmarshal Transaction", func(t *testing.T) {
		tx := &test.TestTransaction
		bin, err := mrsh.MarshalStruct(tx)
		req.NoError(err)
		fmt.Println("Serialized Transaction size: ", len(bin))
		fmt.Printf("%02x\n", bin)
		tx2 := &Transaction{}
		err = mrsh.UnmarshalStruct(bin, tx2)
		req.NoError(err)
		req.Equal(tx, tx2)
	})

	t.Run("Marshal & unmarshal Transaction tuple", func(t *testing.T) {
		tx := &test.TestTransaction
		b1, err := ipldcbor.DumpObject(tx.Type)
		req.NoError(err)
		b2, err := ipldcbor.DumpObject(tx.From)
		req.NoError(err)
		b3, err := ipldcbor.DumpObject(tx.Nonce)
		req.NoError(err)
		b4, err := ipldcbor.DumpObject(tx.To)
		req.NoError(err)
		b5, err := ipldcbor.DumpObject(tx.Data)
		req.NoError(err)
		b6, err := ipldcbor.DumpObject(tx.Sig)
		req.NoError(err)
		l := len(b1) + len(b2) + len(b3) + len(b4) + len(b5) + len(b6)
		bin := make([]byte, 0, len(b1)+len(b2)+len(b3)+len(b4)+len(b5)+len(b6))
		bin = append(append(append(append(append(append(bin, b1...), b2...), b3...), b4...), b5...), b6...)
		req.Equal(l, len(bin))
		fmt.Println("Serialized Transaction size: ", l)
	})

	t.Run("Marshal & unmarshal Transactions", func(t *testing.T) {
		txs := &test.TestTransactions
		bin, err := mrsh.MarshalStructSlice(txs)
		req.NoError(err)
		fmt.Println("Serialized Transactions size: ", len(bin))
		fmt.Printf("%02x\n", bin)
		txs2 := &Transactions{}
		err = mrsh.UnmarshalStructSlice(bin, txs2)
		req.NoError(err)
		req.Equal(txs, txs2)
	})

	//t.Run("Test that Transaction with nil Sig field serialize correctly", func(t *testing.T) {
	//	tx := &test.TestTransaction
	//	tx.Sig = nil
	//	tx2 := &test.TransactionNoSig{
	//		Type:  tx.Type,
	//		From:  tx.From,
	//		Nonce: tx.Nonce,
	//		To:    tx.To,
	//		Data:  tx.Data,
	//		Extra: tx.Extra,
	//	}
	//	bin, err := mrsh.MarshalStruct(tx)
	//	req.NoError(err)
	//	bin2, err := mrsh.MarshalStruct(tx2)
	//	req.NoError(err)
	//	fmt.Printf("%0x\n", bin2)
	//	fmt.Printf("%0x\n", bin)
	//	req.Equal(bin2, bin)
	//})

	t.Run("BlockHeader.MarshalCBOR/UnmarshalCBOR", func(t *testing.T) {
		bh := &test.TestBlockHeader
		var b bytes.Buffer
		err := bh.MarshalCBOR(&b)
		req.NoError(err)
		bin := b.Bytes()
		fmt.Println("Serialized BlockHeader size: ", len(bin))
		bh2 := &BlockHeader{}
		err = bh2.UnmarshalCBOR(&b)
		req.NoError(err)
		req.Equal(bh, bh2)
	})

	t.Run("Marshal & unmarshal BlockHeader", func(t *testing.T) {
		bh := &test.TestBlockHeader
		bin, err := mrsh.MarshalStruct(bh)
		fmt.Println("Serialized BlockHeader size: ", len(bin))
		fmt.Printf("%0x\n", bin)
		req.NoError(err)
		bh2 := &BlockHeader{}
		err = mrsh.UnmarshalStruct(bin, bh2)
		req.NoError(err)
		req.Equal(bh, bh2)
	})
}

func BenchmarkMarshalUnmarshal(b *testing.B) {
	mrsh := test.Mrsh

	b.Run("Marshal Transaction map", func(b *testing.B) {
		tx := &test.TestTransaction
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mrsh.MarshalStruct(tx)
		}
	})

	b.Run("Unmarshal Transaction map", func(b *testing.B) {
		tx := &test.TestTransaction
		bin, _ := mrsh.MarshalStruct(tx)
		tx2 := &Transaction{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			mrsh.UnmarshalStruct(bin, tx2)
		}
	})

	b.Run("Marshal Transaction tuple gen", func(b *testing.B) {
		tx := &test.TestTransaction
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var b bytes.Buffer
			tx.MarshalCBOR(&b)
			b.Bytes()
		}
	})

	b.Run("Marshal Transaction tuple append", func(b *testing.B) {
		tx := &test.TestTransaction
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b1, _ := ipldcbor.DumpObject(tx.Type)
			b2, _ := ipldcbor.DumpObject(tx.From)
			b3, _ := ipldcbor.DumpObject(tx.Nonce)
			b4, _ := ipldcbor.DumpObject(tx.To)
			b5, _ := ipldcbor.DumpObject(tx.Data)
			b6, _ := ipldcbor.DumpObject(tx.Sig)
			bin := make([]byte, 0, len(b1)+len(b2)+len(b3)+len(b4)+len(b5)+len(b6))
			bin = append(append(append(append(append(append(bin, b1...), b2...), b3...), b4...), b5...), b6...)
		}
	})

	b.Run("Marshal Transaction tuple copy", func(b *testing.B) {
		tx := &test.TestTransaction
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b1, _ := ipldcbor.DumpObject(tx.Type)
			b2, _ := ipldcbor.DumpObject(tx.From)
			b3, _ := ipldcbor.DumpObject(tx.Nonce)
			b4, _ := ipldcbor.DumpObject(tx.To)
			b5, _ := ipldcbor.DumpObject(tx.Data)
			b6, _ := ipldcbor.DumpObject(tx.Sig)
			bin := make([]byte, 0, len(b1)+len(b2)+len(b3)+len(b4)+len(b5)+len(b6))
			copy(bin, b1)
			copy(bin[len(bin):], b2)
			copy(bin[len(bin):], b3)
			copy(bin[len(bin):], b4)
			copy(bin[len(bin):], b5)
			copy(bin[len(bin):], b6)
		}
	})
}
