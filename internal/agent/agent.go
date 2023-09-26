package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/1eedaegon/distributed-logging-storage-practice/internal/discovery"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Agent struct {
	Config

	log        *log.Log
	server     *grpc.Server
	membership *discovery.Membership
	replicator *log.Replicator

	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

type Config struct {
	ServerTLSConfig *tls.Config
	PeerTLSConfig   *tls.Config
	DataDir         string
	BindAddr        string
	RPCPort         int
	NodeName        string
	StartJoinAddrs  []string
	ACLModelFile    string
	ACLPolicyFile   string
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}
	setup := []func() error{
		a.setupLogger,
		a.setupLog,
		a.setupServer,
		a.setupLog,
	}
	for _, fn := range setup {
		if err := fn(); err != nil {
			return a, err
		}
	}
	return a, nil
}

func (a *Agent) setupLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}
func (a *Agent) setupLog() error {
	return nil
}
func (a *Agent) setupServer() error {
	return nil
}
func (a *Agent) setupMembership() error {
	return nil
}
