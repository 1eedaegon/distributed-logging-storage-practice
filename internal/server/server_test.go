package server

import (
	"testing"

	api "github.com/1eedaegon/distributed-logging-storage-practice/api/v1"
)

func testProduceConsume(t *testing.T, client api.LogClient, config *Config)       {}
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
	return (client, cfg, func() {})
}
