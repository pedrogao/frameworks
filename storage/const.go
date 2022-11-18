package storage

import "errors"

const (
	magicNumberSize = 4
	counterSize     = 4
	nodeHeaderSize  = 3

	collectionSize = 16
	pageNumSize    = 8
)

var WriteInsideReadTxErr = errors.New("can't perform a write operation inside a read transaction")
