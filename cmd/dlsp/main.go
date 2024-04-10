package main

import (
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/1eedaegon/distributed-logging-storage-practice/internal/agent"
	"github.com/1eedaegon/distributed-logging-storage-practice/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type cli struct {
	cfg cfg
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	var err error

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	c.cfg.DataDir = viper.GetString("data-dir")
	c.cfg.NodeName = viper.GetString("node-name")
	c.cfg.BindAddr = viper.GetString("bind-addr")
	c.cfg.RPCPort = viper.GetInt("rpc-port")
	c.cfg.StartJoinAddrs = viper.GetStringSlice("start-join-addrs")
	c.cfg.Bootstrap = viper.GetBool("bootstrap")
	c.cfg.ACLModelFile = viper.GetString("acl-model-file")
	c.cfg.ACLPolicyFile = viper.GetString("acl-policy-file")
	c.cfg.PeerTLSConfig.CAFile = viper.GetString("peer-tls-ca-file")
	c.cfg.PeerTLSConfig.KeyFile = viper.GetString("peer-tls-key-file")
	c.cfg.PeerTLSConfig.CertFile = viper.GetString("peer-tls-cert-file")
	c.cfg.ServerTLSConfig.CAFile = viper.GetString("server-tls-ca-file")
	c.cfg.ServerTLSConfig.KeyFile = viper.GetString("server-tls-key-file")
	c.cfg.ServerTLSConfig.CertFile = viper.GetString("server-tls-cert-file")

	if c.cfg.ServerTLSConfig.KeyFile != "" && c.cfg.ServerTLSConfig.CertFile != "" {
		c.cfg.ServerTLSConfig.Server = true
		c.cfg.Config.ServerTLSConfig, err = config.SetupTLSConfig(c.cfg.ServerTLSConfig)
		if err != nil {
			return err
		}
	}
	if c.cfg.PeerTLSConfig.KeyFile != "" && c.cfg.PeerTLSConfig.CertFile != "" {
		c.cfg.Config.PeerTLSConfig, err = config.SetupTLSConfig(c.cfg.PeerTLSConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	var err error

	agent, err := agent.New(c.cfg.Config)
	if err != nil {
		return err
	}

	// Graceful shutdown from systemcall
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	return agent.Shutdown()
}

type cfg struct {
	agent.Config
	ServerTLSConfig config.TLSConfig
	PeerTLSConfig   config.TLSConfig
}

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use:     "dslp",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	if err := setupFlags(cmd); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func setupFlags(cmd *cobra.Command) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	cmd.Flags().String("config-file", "", "Path to config file.")
	dataDir := path.Join(os.TempDir(), "dlsp")
	cmd.Flags().String("data-dir", dataDir, "Directory to store log and Raft data.")
	cmd.Flags().String("node-name", hostname, "A unique server ID.")
	cmd.Flags().String("bind-addr", "127.0.0.1:8401", "Address to bind Serf on it.")
	cmd.Flags().Int("rpc-port", 8400, "Port for RPC Clients(Raft) connections.")
	cmd.Flags().StringSlice("start-join-addrs", nil, "Serf addresses to join.")
	cmd.Flags().Bool("bootstrap", false, "Bootstrap for cluster.")
	cmd.Flags().String("acl-model-file", "", "Path to ACL model file")
	cmd.Flags().String("acl-policy-file", "", "Path to ACL policy file")
	cmd.Flags().String("server-tls-ca-file", "", "Path to server certificate authority(CA) file")
	cmd.Flags().String("server-tls-key-file", "", "Path to server TLS Key file")
	cmd.Flags().String("server-tls-cert-file", "", "Path to server TLS Cert file")
	cmd.Flags().String("peer-tls-ca-file", "", "Path to peer TLS certificate authority(CA) file")
	cmd.Flags().String("peer-tls-key-file", "", "Path to peer TLS Key file")
	cmd.Flags().String("peer-tls-cert-file", "", "Path to peer TLS Cert file")

	return viper.BindPFlags(cmd.Flags())
}
