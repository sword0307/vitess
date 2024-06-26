/*
Copyright 2019 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package grpcbinlogplayer

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"

	"vitess.io/vitess/go/vt/binlog/binlogplayertest"
	"vitess.io/vitess/go/vt/binlog/grpcbinlogstreamer"

	binlogservicepb "vitess.io/vitess/go/vt/proto/binlogservice"
	topodatapb "vitess.io/vitess/go/vt/proto/topodata"
)

// the test here creates a fake server implementation, a fake client
// implementation, and runs the test suite against the setup.
func TestGRPCBinlogStreamer(t *testing.T) {
	// Listen on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Cannot listen: %v", err)
	}
	host := listener.Addr().(*net.TCPAddr).IP.String()
	port := listener.Addr().(*net.TCPAddr).Port

	// Create a gRPC server and listen on the port
	server := grpc.NewServer()
	fakeUpdateStream := binlogplayertest.NewFakeBinlogStreamer(t)
	binlogservicepb.RegisterUpdateStreamServer(server, grpcbinlogstreamer.New(fakeUpdateStream))
	go server.Serve(listener)

	// Create a GRPC client to talk to the fake tablet
	c := &client{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// and send it to the test suite
	binlogplayertest.Run(ctx, t, c, &topodatapb.Tablet{
		Hostname: host,
		PortMap: map[string]int32{
			"grpc": int32(port),
		},
	}, fakeUpdateStream)
}
