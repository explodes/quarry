package userservice

import (
	"context"

	"github.com/explodes/quarry/examples/rpcd/rpcdpb"
	"google.golang.org/grpc"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/rpcd/rpcdquarry"
)

func init() {
	q := rpcdquarry.Default()

	q.MustAddFactory("userService", quarry.Singleton(buildUserService))

	q.MustAddFactory("registerUserService", quarry.Singleton(registerUserService))
	q.MustAddDependency("registerUserService", "grpcServer")
	q.MustAddDependency("registerUserService", "userService")

	q.MustAddFactory("userdDialOptions", quarry.Provider([]grpc.DialOption{grpc.WithInsecure()}))

	q.MustAddFactory("userdClientConn", quarry.Singleton(buildUserdClientConn))
	q.MustAddDependency("userdClientConn", "userdAddress")
	q.MustAddDependency("userdClientConn", "userdDialOptions")

	q.MustAddFactory("userdClient", quarry.Singleton(buildUserdClient))
	q.MustAddDependency("userdClient", "userdClientConn")
}

func buildUserService(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	return &userService{}, nil
}

func registerUserService(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	grpcServer := deps["grpcServer"].(*grpc.Server)
	userService := deps["userService"].(*userService)

	rpcdpb.RegisterUserServiceServer(grpcServer, userService)
	return nil, nil
}

func buildUserdClientConn(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	address := deps["userdAddress"].(string)
	userdDialOptions := deps["userdDialOptions"].([]grpc.DialOption)

	return grpc.DialContext(ctx, address, userdDialOptions...)
}

func buildUserdClient(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	userdClientConn := deps["userdClientConn"].(*grpc.ClientConn)

	return rpcdpb.NewUserServiceClient(userdClientConn), nil
}
