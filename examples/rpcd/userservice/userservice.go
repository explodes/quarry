package userservice

import (
	"log"

	"github.com/explodes/quarry/examples/rpcd/rpcdpb"
	"github.com/explodes/quarry/examples/rpcd/rpcdquarry"
	"golang.org/x/net/context"
)

var _ rpcdpb.UserServiceServer = (*userService)(nil)

type userService struct {
}

func (u *userService) fetchFromQuarry(ctx context.Context, request interface{}, name string) (interface{}, error) {
	return rpcdquarry.Default().Get(ctx, request, name)
}

func (u *userService) CreateUser(ctx context.Context, request *rpcdpb.CreateUserRequest) (*rpcdpb.CreateUserResponse, error) {
	log.Println("CreateUser")
	user, err := u.fetchFromQuarry(ctx, request, "createUser")
	if err != nil {
		return nil, err
	}
	response := &rpcdpb.CreateUserResponse{
		User: user.(*rpcdpb.User),
	}
	return response, nil
}

func (u *userService) Login(ctx context.Context, request *rpcdpb.LoginRequest) (*rpcdpb.LoginResponse, error) {
	log.Println("Login")
	token, err := u.fetchFromQuarry(ctx, request, "loginUser")
	if err != nil {
		return nil, err
	}
	response := &rpcdpb.LoginResponse{
		Token: token.(string),
	}
	return response, nil
}

func (u *userService) Validate(ctx context.Context, request *rpcdpb.ValidateRequest) (*rpcdpb.ValidateResponse, error) {
	log.Println("Validate")
	user, err := u.fetchFromQuarry(ctx, request, "user")
	if err != nil {
		return nil, err
	}
	response := &rpcdpb.ValidateResponse{
		User: user.(*rpcdpb.User),
	}
	return response, nil
}
