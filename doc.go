// Copyright (c) 2013-2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

/*
Package rddutil provides Reddcoin-specific convenience functions and types.

Block Overview

A Block defines a Reddcoin block that provides easier and more efficient
manipulation of raw wire protocol blocks.  It also memoizes hashes for the
block and its transactions on their first access so subsequent accesses don't
have to repeat the relatively expensive hashing operations.

Tx Overview

A Tx defines a Reddcoin transaction that provides more efficient manipulation of
raw wire protocol transactions.  It memoizes the hash for the transaction on its
first access so subsequent accesses don't have to repeat the relatively
expensive hashing operations.

Base58 Usage

To decode a base58 string:

 rawData := rddutil.Base58Decode(encodedData)

Similarly, to encode the same data:

 encodedData := rddutil.Base58Encode(rawData)

*/
package rddutil
