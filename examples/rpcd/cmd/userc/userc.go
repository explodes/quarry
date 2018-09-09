package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/explodes/quarry/examples/rpcd/rpcdpb"

	"google.golang.org/grpc"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/rpcd/rpcdquarry"

	_ "github.com/explodes/quarry/examples/rpcd/userservice"
	_ "github.com/explodes/quarry/examples/rpcd/userstorage"
)

const (
	envAddress     = "USERD_ADDRESS"
	defaultAddress = "0.0.0.0:9001"

	envUsername     = "USERC_USERNAME"
	defaultUsername = "explodes"

	envPassword     = "USERC_PASSWORD"
	defaultPassword = "supersecret"

	rpcTimeoutDuration = 10 * time.Second
)

func init() {
	q := rpcdquarry.Default()

	q.MustAddFactory("userdAddress", quarry.Provider(getEnv(envAddress, defaultAddress)))
}

func main() {
	q := rpcdquarry.Default()

	username := getEnv(envUsername, defaultUsername)
	password := getEnvSensitive(envPassword, defaultPassword)

	defer func() {
		userdClientConn, err := q.Get(context.Background(), nil, "userdClientConn")
		if err != nil {
			log.Fatal(err)
		}
		if err := userdClientConn.(*grpc.ClientConn).Close(); err != nil {
			log.Fatal(err)
		}
	}()

	userdClientInterface, err := q.Get(context.Background(), nil, "userdClient")
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	client := userdClientInterface.(rpcdpb.UserServiceClient)

	// CreateUser.
	createUserRequest := &rpcdpb.CreateUserRequest{
		Username: username,
		Password: password,
	}
	fmt.Printf("CreateUser -> %s\n", createUserRequest)
	createUserResponse, err := client.CreateUser(rpcTimeout(), createUserRequest)
	printResult("CreateUser", createUserResponse, err)

	// LoginUser.
	loginRequest := &rpcdpb.LoginRequest{
		Username: username,
		Password: password,
	}
	fmt.Printf("LoginUser -> %s\n", loginRequest)
	loginResponse, err := client.Login(rpcTimeout(), loginRequest)
	printResult("LoginUser", loginResponse, err)

	// Validate.
	validateRequest := &rpcdpb.ValidateRequest{
		Token: loginResponse.Token,
	}
	fmt.Printf("Validate -> %s\n", createUserRequest)
	validateResponse, err := client.Validate(rpcTimeout(), validateRequest)
	printResult("Validate", validateResponse, err)
}

// ## Utils ##

func printResult(name string, i interface{}, err error) {
	if err != nil {
		log.Printf("%s <- error creating user: %v", name, err)
	} else {
		fmt.Printf("%s <- %s\n", name, i)
	}
}

func rpcTimeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), rpcTimeoutDuration)
	return ctx
}

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
