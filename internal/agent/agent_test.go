package agent_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	apiv1 "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/agent"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/config"
	port "github.com/1eedaegon/go-dynamic-port-allocator"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func TestAgent(t *testing.T) {

	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	var agents []*agent.Agent
	for i := 0; i < 3; i++ {
		log.Printf("[SPAWNed AGENT 1]\n")
		ports := port.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]

		dataDir, err := os.MkdirTemp("", "agent-test-log")
		require.NoError(t, err)

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}

		agent, err := agent.New(agent.Config{
			NodeName:        fmt.Sprintf("%d", i),
			Bootstrap:       i == 0,
			StartJoinAddrs:  startJoinAddrs,
			BindAddr:        bindAddr,
			RPCPort:         rpcPort,
			DataDir:         dataDir,
			ACLModelFile:    config.ACLModelFile,
			ACLPolicyFile:   config.ACLPolicyFile,
			ServerTLSConfig: serverTLSConfig,
			PeerTLSConfig:   peerTLSConfig,
		})
		require.NoError(t, err)

		agents = append(agents, agent)
	}
	defer func() {
		for _, agent := range agents {
			log.Printf("[SPAWNed AGENT 2]\n")
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t, os.RemoveAll(agent.Config.DataDir))
		}
	}()

	time.Sleep(3 * time.Second)

	leaderClient := client(t, agents[0], peerTLSConfig)
	produceResponse, err := leaderClient.Produce(
		context.Background(),
		&apiv1.ProduceRequest{
			Record: &apiv1.Record{
				Value: []byte("foo"),
			},
		},
	)
	require.NoError(t, err)
	consumeResponse, err := leaderClient.Consume(
		context.Background(),
		&apiv1.ConsumeRequest{
			Offset: produceResponse.Offset,
		},
	)
	require.NoError(t, err)
	require.Equal(t, consumeResponse.Record.Value, []byte("foo"))

	time.Sleep(3 * time.Second)

	// replica에서 데이터가 복제되는지 확인
	followerClient := client(t, agents[1], peerTLSConfig)
	consumeResponse, err = followerClient.Consume(
		context.Background(),
		&apiv1.ConsumeRequest{
			Offset: produceResponse.Offset,
		},
	)
	require.NoError(t, err)
	require.Equal(t, consumeResponse.Record.Value, []byte("foo"))

	// replica에서 데이터가 무한증식하는지 확인(leader/follow를 지정하지 않았을 때)
	consumeResponse, err = leaderClient.Consume(
		context.Background(),
		&apiv1.ConsumeRequest{
			Offset: produceResponse.Offset + 1,
		},
	)
	// 다음 요청이 없어야한다.
	require.Nil(t, consumeResponse)
	// 없기 때문에 ErrOffsetOutOfRange 에러가 나와야한다.
	require.Error(t, err)
	got := status.Code(err)
	want := status.Code(apiv1.ErrOffsetOutOfRange{}.GRPCStatus().Err())
	require.Equal(t, got, want)
}

// Client test helper
func client(t *testing.T, agent *agent.Agent, tlsConfig *tls.Config) apiv1.LogClient {
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)
	conn, err := grpc.Dial(rpcAddr, opts...)
	require.NoError(t, err)
	cli := apiv1.NewLogClient(conn)
	return cli
}
