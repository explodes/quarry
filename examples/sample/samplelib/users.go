package samplelib

import (
	"context"
	"fmt"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/sample/samplepb"
	"github.com/explodes/quarry/examples/sample/samplequarry"
)

func init() {
	graph := samplequarry.Default()

	// Dependencies for UserService would normally be provided by the graph.
	// For a service-like object, consider a quarry.Singleton.
	userService := &UserService{}
	graph.MustAddFactory("userService", quarry.Provider(userService))

	graph.MustAddFactory("user", fetchUser)
	graph.MustAddDependency("user", "userService")
}

type UserService struct{}

func (s *UserService) FetchUserForToken(ctx context.Context, token string) (*samplepb.User, error) {
	fmt.Println("UserService::FetchUserForToken")
	// Simulate using context.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// "Fetch" our authenticated user.
	user := &samplepb.User{
		Username: "taco",
		Email:    "taco@example.com",
	}
	return user, nil
}

func fetchUser(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	request := params.(*samplepb.SampleRequest)
	userService := deps["userService"].(*UserService)

	return userService.FetchUserForToken(ctx, request.Token)
}
