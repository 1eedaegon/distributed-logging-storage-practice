package server

import (
	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
)

type CommitLog interface {
	Append(*api.Record) (uint64, error)
	Read(uint64) (*api.Record, error)
}

type Config struct {
	CommitLog CommitLog
}
