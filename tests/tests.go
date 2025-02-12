package tests

import (
	"context"
	"testing"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/grpc"

	"github.com/openfga/openfga/pkg/testfixtures/storage"

	"github.com/openfga/openfga/pkg/testutils"

	"github.com/openfga/openfga/cmd/run"
	serverconfig "github.com/openfga/openfga/internal/server/config"
	"github.com/openfga/openfga/pkg/logger"
)

// TestClientBootstrapper defines a client interface definition that can be used by tests
// to bootstrap OpenFGA resources (stores, models, relationship tuples, etc..) needed to
// execute tests.
type TestClientBootstrapper interface {
	CreateStore(ctx context.Context, in *openfgav1.CreateStoreRequest, opts ...grpc.CallOption) (*openfgav1.CreateStoreResponse, error)
	WriteAuthorizationModel(ctx context.Context, in *openfgav1.WriteAuthorizationModelRequest, opts ...grpc.CallOption) (*openfgav1.WriteAuthorizationModelResponse, error)
	Write(ctx context.Context, in *openfgav1.WriteRequest, opts ...grpc.CallOption) (*openfgav1.WriteResponse, error)
}

// StartServer calls StartServerWithContext. See the docs for that.
func StartServer(t testing.TB, cfg *serverconfig.Config) {
	logger := logger.MustNewLogger(cfg.Log.Format, cfg.Log.Level, cfg.Log.TimestampFormat)
	serverCtx := &run.ServerContext{Logger: logger}
	StartServerWithContext(t, cfg, serverCtx)
}

// StartServerWithContext starts a server with a specific ServerContext and waits until it is healthy.
// When the test ends, all resources are cleaned.
func StartServerWithContext(t testing.TB, cfg *serverconfig.Config, serverCtx *run.ServerContext) {
	container := storage.RunDatastoreTestContainer(t, cfg.Datastore.Engine)
	cfg.Datastore.URI = container.GetConnectionURI(true)

	ctx, cancel := context.WithCancel(context.Background())

	serverDone := make(chan error)
	go func() {
		serverDone <- serverCtx.Run(ctx, cfg)
	}()
	t.Cleanup(func() {
		t.Log("waiting for server to stop")
		cancel()
		serverErr := <-serverDone
		t.Log("server stopped with error: ", serverErr)
	})

	testutils.EnsureServiceHealthy(t, cfg.GRPC.Addr, cfg.HTTP.Addr, nil, false)
}
