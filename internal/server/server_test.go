package server

import (
	"io/ioutil"
	"net"
	"testing"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/log"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func testProduceConsume(t *testing.T, client api.LogClient, config *Config)       {

}
func testProduceConsumeStream(t *testing.T, client api.LogClient, config *Config) {}
func testConsumePastBoundary(t *testing.T, client api.LogClient, config *Config)  {}

func TestServer(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, client api.LogClient, config *Config){
		"produce/consume a message to/from the log success": testProduceConsume,
		"produce/consume a stream succeeds":                 testProduceConsumeStream,
		"consume past log boundary fails":                   testConsumePastBoundary,
	} {
		t.Run(scenario, func(t *testing.T) {
			client, config, teardown := setupTest(t, nil)
			defer teardown()
			fn(t, client, config)
		})
	}
}

func setupTest(t *testing.T, fn func(*Config)) (client api.LogClient, cfg *Config teardown func()){
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	clientOptions := []grpc.DialOption{grpc.WithInsecure()}
	cc, err := grpc.Dial(l.Addr().String(), clientOptions...)
	require.NoError(t, err)

	dir, err := ioutil.TempDir("", "server-testing-dir")
	require.NoError(t, err)

	clog, err := log.NewLog(dir, log.Config{})
	require.NoError(t, err)

	cfg = &Config{
		CommitLog: clog,
	}
	if fn != nil {
		fn(cfg)
	}
	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func() {
		server.Serve()
	}()
	client = api.NewLogClient(cc)

	return client, cfg, func() {
		server.Stop()
		cc.Close()
		l.Close()
		clog.Remove()
	}
}
