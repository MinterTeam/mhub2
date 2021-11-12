package keeper

import (
	"net/http"

	mhub2Gw "github.com/MinterTeam/mhub2/module/x/mhub2/types"
	oracleGw "github.com/MinterTeam/mhub2/module/x/oracle/types"
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

func RunHttpServer(httpListenAddr, grpcServerAddr string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := oracleGw.RegisterQueryHandlerFromEndpoint(ctx, mux, grpcServerAddr, opts); err != nil {
		return err
	}
	if err := mhub2Gw.RegisterQueryHandlerFromEndpoint(ctx, mux, grpcServerAddr, opts); err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(httpListenAddr, mux)
}
