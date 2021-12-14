package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/daotl/go-doubl/model"
	"github.com/daotl/go-doubl/test"
)

func TestComputeRootHash(t *testing.T) {
	req := require.New(t)
	assr := assert.New(t)
	var err error
	ut := test.Util

	txs := test.TestTransactions
	txhs := make([]TransactionHash, len(txs))
	for i, tx := range txs {
		txhs[i], err = ut.HashTransaction(&tx)
		req.NoError(err)
	}
	h := ut.GenRootHashFromTransactionHashes(txhs)

	h_, err := ut.GenRootHashFromTransactions(txs)
	req.NoError(err)
	assr.Equal(h, h_)

	txxs, err := ut.ExtendTransactions(txs)
	req.NoError(err)
	h_ = ut.GenRootHashFromTransactionExts(txxs)
	assr.Equal(h, h_)
}
