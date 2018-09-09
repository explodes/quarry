package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/rpcd/rpcdquarry"

	_ "github.com/explodes/quarry/examples/rpcd/userservice"
	_ "github.com/explodes/quarry/examples/rpcd/userstorage"
)

const (
	envSecret     = "USERD_SECRET"
	defaultSecret = "supersecret"

	envBind     = "USERD_BIND"
	defaultBind = "0.0.0.0:9001"
)

func init() {
	q := rpcdquarry.Default()

	q.MustAddFactory("secret", quarry.Provider(getEnvSensitive(envSecret, defaultSecret)))

	q.MustAddFactory("bind", quarry.Provider(getEnv(envBind, defaultBind)))

	q.MustAddFactory("userdListener", quarry.Singleton(buildUserdListener))
	q.MustAddDependency("userdListener", "bind")

	q.MustAddFactory("grpcServerOptions", quarry.Provider([]grpc.ServerOption(nil)))

	q.MustAddFactory("grpcServer", quarry.Singleton(buildGrpcServer))
	q.MustAddDependency("grpcServer", "grpcServerOptions")

	q.MustAddFactory("userdRunner", quarry.Singleton(buildUserdRunner))
	q.MustAddDependency("userdRunner", "registerUserService")
	q.MustAddDependency("userdRunner", "userdListener")
	q.MustAddDependency("userdRunner", "grpcServer")
}

func main() {
	q := rpcdquarry.Default()

	userdRunner, err := q.Get(context.Background(), nil, "userdRunner")
	if err != nil {
		log.Fatalf("error creating server: %v", err)
	}
	if err := userdRunner.(func() error)(); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func buildUserdListener(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	bind := deps["bind"].(string)

	log.Printf("listening on %s", bind)

	return net.Listen("tcp", bind)
}

func buildUserdRunner(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	userdListener := deps["userdListener"].(net.Listener)
	grpcServer := deps["grpcServer"].(*grpc.Server)

	userdRunner := func() error {
		log.Printf("starting grpc server")
		return grpcServer.Serve(userdListener)
	}
	return userdRunner, nil
}

func buildGrpcServer(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	grpcServerOptions := deps["grpcServerOptions"].([]grpc.ServerOption)

	return grpc.NewServer(grpcServerOptions...), nil
}

// ## Utils ##

func getEnvSensitive(env, defaultValue string) string {
	s, ok := os.LookupEnv(env)
	if !ok {
		log.Printf("Using default value for %s", env)
		return defaultValue
	}
	return s
}

func getEnv(env, defaultValue string) string {
	s, ok := os.LookupEnv(env)
	if !ok {
		log.Printf("Using default value for %s=%s", env, defaultValue)
		return defaultValue
	}
	return s
}
