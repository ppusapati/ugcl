package grpc

import (
	"io"
	"net"
	"strconv"

	"p9e.in/ugcl/packages/config"
	"p9e.in/ugcl/packages/middleware/dbmiddleware"
	"p9e.in/ugcl/packages/middleware/localize"
	"p9e.in/ugcl/packages/middleware/tenant"
	"p9e.in/ugcl/packages/p9log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GRPCServer interface {
	Start(serviceRegister func(server *grpc.Server))
	io.Closer
}

type gRPCServer struct {
	grpcServer *grpc.Server
	config     config.GrpcServerConfig
	log        p9log.Helper
}

func NewGrpcServer(config config.GrpcServerConfig, includeTenant bool, log p9log.Helper) (GRPCServer, error) {
	options, err := buildOptions(config, includeTenant)
	if err != nil {
		return nil, err
	}

	server := grpc.NewServer(options...)

	return &gRPCServer{
		config:     config,
		grpcServer: server,
		log:        log,
	}, err
}

func buildOptions(config config.GrpcServerConfig, includeTenant bool) ([]grpc.ServerOption, error) {
	interceptors := []grpc.UnaryServerInterceptor{
		localize.I18N,
		dbmiddleware.NewDBResolver(config.DBContext).DbMiddleware,
	}
	if includeTenant {
		interceptors = append(interceptors, tenant.GrpcTenantMiddleware)
	}

	return []grpc.ServerOption{
		grpc.KeepaliveParams(buildKeepaliveParams(config.KeepaliveParams)),
		grpc.KeepaliveEnforcementPolicy(buildKeepalivePolicy(config.KeepalivePolicy)),
		grpc.ChainUnaryInterceptor(interceptors...),
	}, nil
}

func buildKeepalivePolicy(config keepalive.EnforcementPolicy) keepalive.EnforcementPolicy {
	return keepalive.EnforcementPolicy{
		MinTime:             config.MinTime,
		PermitWithoutStream: config.PermitWithoutStream,
	}
}

func buildKeepaliveParams(config keepalive.ServerParameters) keepalive.ServerParameters {
	return keepalive.ServerParameters{
		MaxConnectionIdle:     config.MaxConnectionIdle,
		MaxConnectionAge:      config.MaxConnectionAge,
		MaxConnectionAgeGrace: config.MaxConnectionAgeGrace,
		Time:                  config.Time,
		Timeout:               config.Timeout,
	}
}

func (g gRPCServer) Start(serviceRegister func(server *grpc.Server)) {
	grpcListener, err := net.Listen("tcp", ":"+strconv.Itoa(int(g.config.Port)))
	if err != nil {
		g.log.Error("failed to start grpc server", err)
	}

	serviceRegister(g.grpcServer)

	g.log.Info("start grpc server success, Endpoint: ", grpcListener.Addr())
	if err := g.grpcServer.Serve(grpcListener); err != nil {
		g.log.Error("failed to grpc server serve", err)
	}
}

func (g gRPCServer) Close() error {
	g.log.Info("close gRPC server")
	g.grpcServer.GracefulStop()
	return nil
}
