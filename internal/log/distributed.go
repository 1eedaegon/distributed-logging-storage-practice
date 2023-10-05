package log

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"go.starlark.net/lib/proto"
)

type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}

func NewDistributedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{config: config}
	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}
	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := filepath.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	var err error
	l.log, err = NewLog(logDir, l.config)
	return err
}

/*
== RAFT ==
- Finite state machine: Service status
- Log store
- Snapshot store
- Stable store: Service metadata(BoltDB)
- Transport
*/
func (l *DistributedLog) setupRaft(dataDir string) error {
	fsm := &fsm{log: l.log}
	logDir := filepath.Join(dataDir, "raft", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	logConfig := l.config
	logConfig.Segment.InitialOffset = 1
	logStore, err := newLogStore()

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable"))
	if err != nil {
		return err
	}

	retain := 1 // 스냅샷 개수
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"),
		retain,
		os.Stderr,
	)
	if err != nil {
		return err
	}

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		l.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)
	config := raft.DefaultConfig()
	config.LocalID = l.config.Raft.LocalID
	if l.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = l.config.Raft.HeartbeatTimeout
	}
	if l.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = l.config.Raft.HeartbeatTimeout
	}
	if l.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
	}
	if l.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = l.config.Raft.CommitTimeout
	}

	l.raft, err = raft.NewRaft(
		config,
		fsm,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)

	if err != nil {
		return err
	}

	if l.config.Raft.Bootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			}},
		}
		err = l.raft.BootstrapCluster(config).Error()
	}
	return err
}

// Produce의 수행 주체는 raft
func (l *DistributedLog) Append(record *api.Record) (uint64, error) {
	res, err := l.apply(
		AppendRequestType,
		&api.ProduceRequest{Record: record},
	)
	if err != nil {
		return 0, err
	}
	return res.(*api.ProduceResponse).Offset, nil
}

func (l *DistributedLog) apply(reqType RequestType, req proto.Message) (interface{}, error) {
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}

	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}
	timeout := 10 * time.Second
	future := l.raft.Apply(buf.Bytes(), timeout)
	if future.Error() != nil {
		return nil, future.Error()
	}
	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}

	return res, nil
}

// Relaxed consistency: raft를 거치지 않고 읽는다.
// 만약 strong consistency를 구현하고 싶다면 raft를 거쳐서 읽는다.
// 읽는 시간이 늘어난다.
func (l *DistributedLog) Read(offset uint64) (*api.Record, error) {
	return l.log.Read(offset)
}

var _ raft.FSM = (*fsm)(nil)

type fsm struct {
	log *Log
}

type RequestType uint8

const (
	AppendRequestType RequestType = 0
)

func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])
	switch reqType {
	case AppendRequestType:
		return l.applyAppend(buf[1:])
	}
	return nil
}

func (l *fsm) applyAppend(b []byte) interface{} {
	var req api.ProduceRequest
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}
	offset, err := l.log.Append(req.Record)
	if err != nil {
		return err
	}
	return &api.ProduceResponse{Offset: offset}
}
