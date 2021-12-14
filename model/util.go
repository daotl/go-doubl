package model

// TransactionsFromExts extracts []Transaction from []TransactionExt
func TransactionsFromExts(txxs []TransactionExt) []Transaction {
	txs := make([]Transaction, len(txxs))
	for _, bhx := range txxs {
		txs = append(txs, *bhx.Transaction)
	}
	return txs
}

// BlockHeadersFromExts extracts []BlockHeader from []BlockHeaderExt
func BlockHeadersFromExts(bhxs []BlockHeaderExt) []BlockHeader {
	bhs := make([]BlockHeader, len(bhxs))
	for _, bhx := range bhxs {
		bhs = append(bhs, *bhx.BlockHeader)
	}
	return bhs
}

// BlocksFromExts extracts []Block from []BlockExt
func BlocksFromExts(bxs []BlockExt) []Block {
	bs := make([]Block, len(bxs))
	for _, bhx := range bxs {
		bs = append(bs, *bhx.Block)
	}
	return bs
}
