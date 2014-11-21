// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rddutil

import (
	"bytes"
	"fmt"
	"io"

	"github.com/reddcoin-project/rddwire"
)

// OutOfRangeError describes an error due to accessing an element that is out
// of range.
type OutOfRangeError string

// BlockHeightUnknown is the value returned for a block height that is unknown.
// This is typically because the block has not been inserted into the main chain
// yet.
const BlockHeightUnknown = int64(-1)

// Error satisfies the error interface and prints human-readable errors.
func (e OutOfRangeError) Error() string {
	return string(e)
}

// Block defines a Reddcoin block that provides easier and more efficient
// manipulation of raw blocks.  It also memoizes hashes for the block and its
// transactions on their first access so subsequent accesses don't have to
// repeat the relatively expensive hashing operations.
type Block struct {
	msgBlock        *rddwire.MsgBlock // Underlying MsgBlock
	serializedBlock []byte            // Serialized bytes for the block
	blockSha        *rddwire.ShaHash  // Cached block hash
	blockHeight     int64             // Height in the main block chain
	transactions    []*Tx             // Transactions
	txnsGenerated   bool              // ALL wrapped transactions generated
}

// MsgBlock returns the underlying rddwire.MsgBlock for the Block.
func (b *Block) MsgBlock() *rddwire.MsgBlock {
	// Return the cached block.
	return b.msgBlock
}

// Bytes returns the serialized bytes for the Block.  This is equivalent to
// calling Serialize on the underlying rddwire.MsgBlock, however it caches the
// result so subsequent calls are more efficient.
func (b *Block) Bytes() ([]byte, error) {
	// Return the cached serialized bytes if it has already been generated.
	if len(b.serializedBlock) != 0 {
		return b.serializedBlock, nil
	}

	// Serialize the MsgBlock.
	var w bytes.Buffer
	err := b.msgBlock.Serialize(&w)
	if err != nil {
		return nil, err
	}
	serializedBlock := w.Bytes()

	// Cache the serialized bytes and return them.
	b.serializedBlock = serializedBlock
	return serializedBlock, nil
}

// Sha returns the block identifier hash for the Block.  This is equivalent to
// calling BlockSha on the underlying rddwire.MsgBlock, however it caches the
// result so subsequent calls are more efficient.
func (b *Block) Sha() (*rddwire.ShaHash, error) {
	// Return the cached block hash if it has already been generated.
	if b.blockSha != nil {
		return b.blockSha, nil
	}

	// Generate the block hash.  Ignore the error since BlockSha can't
	// currently fail.
	sha, _ := b.msgBlock.BlockSha()

	// Cache the block hash and return it.
	b.blockSha = &sha
	return &sha, nil
}

// Tx returns a wrapped transaction (rddutil.Tx) for the transaction at the
// specified index in the Block.  The supplied index is 0 based.  That is to
// say, the first transaction in the block is txNum 0.  This is nearly
// equivalent to accessing the raw transaction (rddwire.MsgTx) from the
// underlying rddwire.MsgBlock, however the wrapped transaction has some helpful
// properties such as caching the hash so subsequent calls are more efficient.
func (b *Block) Tx(txNum int) (*Tx, error) {
	// Ensure the requested transaction is in range.
	numTx := uint64(len(b.msgBlock.Transactions))
	if txNum < 0 || uint64(txNum) > numTx {
		str := fmt.Sprintf("transaction index %d is out of range - max %d",
			txNum, numTx-1)
		return nil, OutOfRangeError(str)
	}

	// Generate slice to hold all of the wrapped transactions if needed.
	if len(b.transactions) == 0 {
		b.transactions = make([]*Tx, numTx)
	}

	// Return the wrapped transaction if it has already been generated.
	if b.transactions[txNum] != nil {
		return b.transactions[txNum], nil
	}

	// Generate and cache the wrapped transaction and return it.
	newTx := NewTx(b.msgBlock.Transactions[txNum])
	newTx.SetIndex(txNum)
	b.transactions[txNum] = newTx
	return newTx, nil
}

// Transactions returns a slice of wrapped transactions (rddutil.Tx) for all
// transactions in the Block.  This is nearly equivalent to accessing the raw
// transactions (rddwire.MsgTx) in the underlying rddwire.MsgBlock, however it
// instead provides easy access to wrapped versions (rddutil.Tx) of them.
func (b *Block) Transactions() []*Tx {
	// Return transactions if they have ALL already been generated.  This
	// flag is necessary because the wrapped transactions are lazily
	// generated in a sparse fashion.
	if b.txnsGenerated {
		return b.transactions
	}

	// Generate slice to hold all of the wrapped transactions if needed.
	if len(b.transactions) == 0 {
		b.transactions = make([]*Tx, len(b.msgBlock.Transactions))
	}

	// Generate and cache the wrapped transactions for all that haven't
	// already been done.
	for i, tx := range b.transactions {
		if tx == nil {
			newTx := NewTx(b.msgBlock.Transactions[i])
			newTx.SetIndex(i)
			b.transactions[i] = newTx
		}
	}

	b.txnsGenerated = true
	return b.transactions
}

// TxSha returns the hash for the requested transaction number in the Block.
// The supplied index is 0 based.  That is to say, the first transaction in the
// block is txNum 0.  This is equivalent to calling TxSha on the underlying
// rddwire.MsgTx, however it caches the result so subsequent calls are more
// efficient.
func (b *Block) TxSha(txNum int) (*rddwire.ShaHash, error) {
	// Attempt to get a wrapped transaction for the specified index.  It
	// will be created lazily if needed or simply return the cached version
	// if it has already been generated.
	tx, err := b.Tx(txNum)
	if err != nil {
		return nil, err
	}

	// Defer to the wrapped transaction which will return the cached hash if
	// it has already been generated.
	return tx.Sha(), nil
}

// TxLoc returns the offsets and lengths of each transaction in a raw block.
// It is used to allow fast indexing into transactions within the raw byte
// stream.
func (b *Block) TxLoc() ([]rddwire.TxLoc, error) {
	rawMsg, err := b.Bytes()
	if err != nil {
		return nil, err
	}
	rbuf := bytes.NewBuffer(rawMsg)

	var mblock rddwire.MsgBlock
	txLocs, err := mblock.DeserializeTxLoc(rbuf)
	if err != nil {
		return nil, err
	}
	return txLocs, err
}

// Height returns the saved height of the block in the block chain.  This value
// will be BlockHeightUnknown if it hasn't already explicitly been set.
func (b *Block) Height() int64 {
	return b.blockHeight
}

// SetHeight sets the height of the block in the block chain.
func (b *Block) SetHeight(height int64) {
	b.blockHeight = height
}

// NewBlock returns a new instance of a Reddcoin block given an underlying
// rddwire.MsgBlock.  See Block.
func NewBlock(msgBlock *rddwire.MsgBlock) *Block {
	return &Block{
		msgBlock:    msgBlock,
		blockHeight: BlockHeightUnknown,
	}
}

// NewBlockFromBytes returns a new instance of a Reddcoin block given the
// serialized bytes.  See Block.
func NewBlockFromBytes(serializedBlock []byte) (*Block, error) {
	br := bytes.NewReader(serializedBlock)
	b, err := NewBlockFromReader(br)
	if err != nil {
		return nil, err
	}
	b.serializedBlock = serializedBlock
	return b, nil
}

// NewBlockFromReader returns a new instance of a Reddcoin block given a
// Reader to deserialize the block.  See Block.
func NewBlockFromReader(r io.Reader) (*Block, error) {
	// Deserialize the bytes into a MsgBlock.
	var msgBlock rddwire.MsgBlock
	err := msgBlock.Deserialize(r)
	if err != nil {
		return nil, err
	}

	b := Block{
		msgBlock:    &msgBlock,
		blockHeight: BlockHeightUnknown,
	}
	return &b, nil
}

// NewBlockFromBlockAndBytes returns a new instance of a Reddcoin block given
// an underlying rddwire.MsgBlock and the serialized bytes for it.  See Block.
func NewBlockFromBlockAndBytes(msgBlock *rddwire.MsgBlock, serializedBlock []byte) *Block {
	return &Block{
		msgBlock:        msgBlock,
		serializedBlock: serializedBlock,
		blockHeight:     BlockHeightUnknown,
	}
}
