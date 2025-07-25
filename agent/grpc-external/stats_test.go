// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package external

import (
	"context"
	"net"
	"sort"
	"testing"

	"github.com/armon/go-metrics"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/hashicorp/go-hclog"

	"github.com/dhiaayachi/consul/agent/consul/rate"
	grpcmiddleware "github.com/dhiaayachi/consul/agent/grpc-middleware"
	"github.com/dhiaayachi/consul/agent/grpc-middleware/testutil"
	"github.com/dhiaayachi/consul/agent/grpc-middleware/testutil/testservice"
	"github.com/dhiaayachi/consul/proto/private/prototest"
)

func TestServer_EmitsStats(t *testing.T) {
	sink, metricsObj := testutil.NewFakeSink(t)

	srv := NewServer(hclog.Default(), metricsObj, nil, rate.NullRequestLimitsHandler(), keepalive.ServerParameters{}, nil)

	testservice.RegisterSimpleServer(srv, &testservice.Simple{})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	lis = grpcmiddleware.LabelledListener{Listener: lis, Protocol: grpcmiddleware.ProtocolPlaintext}

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return srv.Serve(lis)
	})
	t.Cleanup(func() {
		srv.Stop()
		if err := g.Wait(); err != nil {
			t.Logf("grpc server error: %v", err)
		}
	})

	//nolint:staticcheck
	conn, err := grpc.DialContext(ctx, lis.Addr().String(), grpc.WithInsecure())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	client := testservice.NewSimpleClient(conn)
	fClient, err := client.Flow(ctx, &testservice.Req{Datacenter: "mine"})
	require.NoError(t, err)

	// Wait for the first event so that we know the stream is sending.
	_, err = fClient.Recv()
	require.NoError(t, err)

	cancel()
	conn.Close()
	srv.GracefulStop()
	// Wait for the server to stop so that active_streams is predictable.
	require.NoError(t, g.Wait())

	// Occasionally the active_stream=0 metric may be emitted before the
	// active_conns=0 metric. The order of those metrics is not really important
	// so we sort the calls to match the expected.
	sort.Slice(sink.GaugeCalls, func(i, j int) bool {
		if i < 2 || j < 2 {
			return i < j
		}
		if len(sink.GaugeCalls[i].Key) < 4 || len(sink.GaugeCalls[j].Key) < 4 {
			return i < j
		}
		return sink.GaugeCalls[i].Key[3] < sink.GaugeCalls[j].Key[3]
	})

	cmpMetricCalls := cmp.AllowUnexported(testutil.MetricCall{})
	expLabels := []metrics.Label{{
		Name:  "server_type",
		Value: "external",
	}}
	expectedGauge := []testutil.MetricCall{
		{Key: []string{"testing", "grpc", "server", "connections"}, Val: 1, Labels: expLabels},
		{Key: []string{"testing", "grpc", "server", "streams"}, Val: 1, Labels: expLabels},
		{Key: []string{"testing", "grpc", "server", "connections"}, Val: 0, Labels: expLabels},
		{Key: []string{"testing", "grpc", "server", "streams"}, Val: 0, Labels: expLabels},
	}
	prototest.AssertDeepEqual(t, expectedGauge, sink.GaugeCalls, cmpMetricCalls)

	expectedCounter := []testutil.MetricCall{
		{Key: []string{"testing", "grpc", "server", "connection", "count"}, Val: 1, Labels: expLabels},
		{Key: []string{"testing", "grpc", "server", "request", "count"}, Val: 1, Labels: expLabels},
		{Key: []string{"testing", "grpc", "server", "stream", "count"}, Val: 1, Labels: expLabels},
	}
	prototest.AssertDeepEqual(t, expectedCounter, sink.IncrCounterCalls, cmpMetricCalls)
}

func logError(t *testing.T, f func() error) func() {
	return func() {
		if err := f(); err != nil {
			t.Logf(err.Error())
		}
	}
}
