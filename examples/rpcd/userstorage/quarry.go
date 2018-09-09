package userstorage

import (
	"context"

	"github.com/explodes/quarry"
	"github.com/explodes/quarry/examples/rpcd/rpcdquarry"
)

func init() {
	q := rpcdquarry.Default()

	q.MustAddFactory("userStorage", quarry.Singleton(buildUserStorage))
	q.MustAddDependency("userStorage", "secret")

	q.MustAddFactory("user", fetchUser)
	q.MustAddDependency("user", "userStorage")

	q.MustAddFactory("createUser", createUser)
	q.MustAddDependency("createUser", "userStorage")

	q.MustAddFactory("loginUser", loginUser)
	q.MustAddDependency("loginUser", "userStorage")
}

func buildUserStorage(ctx context.Context, deps quarry.Dependencies) (interface{}, error) {
	secret := deps["secret"].(string)

	return newUserStorage(ctx, secret)
}

func fetchUser(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	type hasToken interface {
		GetToken() string
	}
	request := params.(hasToken)
	userStorage := deps["userStorage"].(UserStorage)

	return userStorage.GetUserForToken(ctx, request.GetToken())
}

func createUser(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	type hasSignup interface {
		GetUsername() string
		GetPassword() string
	}
	request := params.(hasSignup)
	userStorage := deps["userStorage"].(UserStorage)

	return userStorage.CreateUser(ctx, request.GetUsername(), request.GetPassword())
}

func loginUser(ctx context.Context, params interface{}, deps quarry.Dependencies) (interface{}, error) {
	type hasLogin interface {
		GetUsername() string
		GetPassword() string
	}
	request := params.(hasLogin)
	userStorage := deps["userStorage"].(UserStorage)

	return userStorage.Login(ctx, request.GetUsername(), request.GetPassword())
}
