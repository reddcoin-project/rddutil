// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rddutil

import (
	"bytes"
	"io"

	"github.com/reddcoin-project/rddwire"
)

// TxIndexUnknown is the value returned for a transaction index that is unknown.
// This is typically because the transaction has not been inserted into a block
// yet.
const TxIndexUnknown = -1

// Tx defines a Reddcoin transaction that provides easier and more efficient
// manipulation of raw transactions.  It also memoizes the hash for the
// transaction on its first access so subsequent accesses don't have to repeat
// the relatively expensive hashing operations.
type Tx struct {
	msgTx   *rddwire.MsgTx   // Underlying MsgTx
	txSha   *rddwire.ShaHash // Cached transaction hash
	txIndex int              // Position within a block or TxIndexUnknown
}

// MsgTx returns the underlying rddwire.MsgTx for the transaction.
func (t *Tx) MsgTx() *rddwire.MsgTx {
	// Return the cached transaction.
	return t.msgTx
}

// Sha returns the hash of the transaction.  This is equivalent to
// calling TxSha on the underlying rddwire.MsgTx, however it caches the
// result so subsequent calls are more efficient.
func (t *Tx) Sha() *rddwire.ShaHash {
	// Return the cached hash if it has already been generated.
	if t.txSha != nil {
		return t.txSha
	}

	// Generate the transaction hash.  Ignore the error since TxSha can't
	// currently fail.
	sha, _ := t.msgTx.TxSha()

	// Cache the hash and return it.
	t.txSha = &sha
	return &sha
}

// Index returns the saved index of the transaction within a block.  This value
// will be TxIndexUnknown if it hasn't already explicitly been set.
func (t *Tx) Index() int {
	return t.txIndex
}

// SetIndex sets the index of the transaction in within a block.
func (t *Tx) SetIndex(index int) {
	t.txIndex = index
}

// NewTx returns a new instance of a Reddcoin transaction given an underlying
// rddwire.MsgTx.  See Tx.
func NewTx(msgTx *rddwire.MsgTx) *Tx {
	return &Tx{
		msgTx:   msgTx,
		txIndex: TxIndexUnknown,
	}
}

// NewTxFromBytes returns a new instance of a Reddcoin transaction given the
// serialized bytes.  See Tx.
func NewTxFromBytes(serializedTx []byte) (*Tx, error) {
	br := bytes.NewReader(serializedTx)
	return NewTxFromReader(br)
}

// NewTxFromReader returns a new instance of a Reddcoin transaction given a
// Reader to deserialize the transaction.  See Tx.
func NewTxFromReader(r io.Reader) (*Tx, error) {
	// Deserialize the bytes into a MsgTx.
	var msgTx rddwire.MsgTx
	err := msgTx.Deserialize(r)
	if err != nil {
		return nil, err
	}

	t := Tx{
		msgTx:   &msgTx,
		txIndex: TxIndexUnknown,
	}
	return &t, nil
}
