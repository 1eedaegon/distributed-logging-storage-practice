package log_test

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	apiv1 "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/log"
	port "github.com/1eedaegon/go-dynamic-port-allocator"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
)

func TestMultipleNodes(t *testing.T) {
	var logs []*log.DistributedLog
	// 분산 노드 3개로 테스트
	nodeCount := 3
	ports := port.Get(nodeCount)

	for i := 0; i < nodeCount; i++ {
		dataDir, err := os.MkdirTemp("", "distributed-log-test")
		require.NoError(t, err)
		defer func(dir string) {
			_ = os.RemoveAll(dir)
		}(dataDir)
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ports[i]))
		require.NoError(t, err)

		config := log.Config{}
		config.Raft.StreamLayer = log.NewStreamLayer(ln, nil, nil)
		config.Raft.LocalID = raft.ServerID(fmt.Sprintf("%d", i))
		config.Raft.HeartbeatTimeout = 50 * time.Millisecond
		config.Raft.ElectionTimeout = 50 * time.Millisecond
		config.Raft.LeaderLeaseTimeout = 50 * time.Millisecond
		config.Raft.CommitTimeout = 50 * time.Millisecond
		// 첫번째 노드는 기본 설정
		if i == 0 {
			config.Raft.Bootstrap = true
		}
		l, err := log.NewDistributedLog(dataDir, config)
		require.NoError(t, err)

		// 첫번째 노드를 제외한 나머지 노드는 첫번째 노드로 join한다.
		// 첫번째 노드는 Leader election을 수행하도록 기다린다.
		if i != 0 {
			err = logs[0].Join(fmt.Sprintf("%d", i), ln.Addr().String())
			require.NoError(t, err)
		} else {
			err = l.WaitForLeader(3 * time.Second)
			require.NoError(t, err)
		}
		logs = append(logs, l)
	}
	records := []*apiv1.Record{
		{Value: []byte("First")},
		{Value: []byte("Second")},
		{Value: []byte("Third")},
	}
	for _, record := range records {
		off, err := logs[0].Append(record)
		require.NoError(t, err)

	}
}