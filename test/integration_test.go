//go:build integration

package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/livekit/protocol/egress"
	"github.com/livekit/protocol/redis"
)

func TestEgress(t *testing.T) {
	conf := NewTestContext(t)

	// rpc client and server
	rc, err := redis.GetRedisClient(conf.Redis)
	require.NoError(t, err)
	rpcServer := egress.NewRedisRPCServer(rc)
	rpcClient := egress.NewRedisRPCClient("egress_test", rc)

	RunTestSuite(t, conf, rpcClient, rpcServer)
}
