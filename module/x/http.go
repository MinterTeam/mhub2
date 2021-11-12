package x

import (
	"context" // Use "golang.org/x/net/context" for Golang version <= 1.6
	"github.com/gorilla/handlers"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	gw1 "github.com/MinterTeam/mhub2/module/x/mhub2/types"
	gw2 "github.com/MinterTeam/mhub2/module/x/oracle/types"
)

func Run(httpPort, grpcServerEndpoint string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux1 := runtime.NewServeMux()
	mux2 := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := gw1.RegisterQueryHandlerFromEndpoint(ctx, mux1, grpcServerEndpoint, opts)
	err = gw2.RegisterQueryHandlerFromEndpoint(ctx, mux2, grpcServerEndpoint, opts)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mhub2/", func(writer http.ResponseWriter, request *http.Request) {
		http.StripPrefix("/mhub2", handlers.CompressHandler(allowCORS(wsproxy.WebsocketProxy(mux1)))).ServeHTTP(writer, request)
	})
	mux.HandleFunc("/oracle/", func(writer http.ResponseWriter, request *http.Request) {
		http.StripPrefix("/oracle", handlers.CompressHandler(allowCORS(wsproxy.WebsocketProxy(mux2)))).ServeHTTP(writer, request)
	})
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(httpPort, mux)
}

func allowCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				preflightHandler(w, r)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func preflightHandler(w http.ResponseWriter, _ *http.Request) {
	headers := []string{"Content-Type", "Accept"}
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
}
