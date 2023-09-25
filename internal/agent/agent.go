package agent

import (
	"crypto/tls"
	"sync"

	"github.com/1eedaegon/distributed-logging-storage-practice/internal/discovery"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/log"
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
