package userstorage

import (
	"context"
	"errors"
	"math/rand"

	"golang.org/x/crypto/bcrypt"

	"github.com/explodes/quarry/examples/rpcd/rpcdpb"
)

const (
	passwordHashCost = 14
	tokenLength      = 64
)

var (
	errInvalidLogin = errors.New("invalid login")
	tokenRunes      = []rune("abcdef0123456789")
)

type UserStorage interface {
	GetUserForToken(ctx context.Context, token string) (*rpcdpb.User, error)
	CreateUser(ctx context.Context, username string, password string) (*rpcdpb.User, error)
	Login(ctx context.Context, username string, password string) (string, error)
}

var _ UserStorage = (*userStorage)(nil)

type userStorage struct {
	secret string

	users  map[string]userRecord
	tokens map[string]string
}

type userRecord struct {
	public       *rpcdpb.User
	passwordHash string
}

func newUserStorage(ctx context.Context, secret string) (*userStorage, error) {
	storage := &userStorage{
		secret: secret,
		users:  make(map[string]userRecord),
		tokens: make(map[string]string),
	}
	return storage, nil
}

func (u *userStorage) GetUserForToken(ctx context.Context, token string) (*rpcdpb.User, error) {
	username, ok := u.tokens[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	user, ok := u.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user.public, nil
}

func (u *userStorage) CreateUser(ctx context.Context, username string, password string) (*rpcdpb.User, error) {
	user, ok := u.users[username]
	if ok {
		return nil, errors.New("username already exists")
	}
	hash, err := u.hashPassword(password)
	if err != nil {
		return nil, err
	}
	user = userRecord{
		public: &rpcdpb.User{
			Username: username,
		},
		passwordHash: hash,
	}
	u.users[username] = user
	return user.public, nil
}

func (u *userStorage) Login(ctx context.Context, username string, password string) (string, error) {
	user, ok := u.users[username]
	if !ok {
		return "", errInvalidLogin
	}
	if !u.checkPassword(password, user.passwordHash) {
		return "", errInvalidLogin
	}
	token := u.randomToken()
	u.tokens[token] = user.public.Username
	return token, nil
}

func (u *userStorage) hashPassword(password string) (string, error) {
	salted := u.saltedPassword(password)
	hash, err := bcrypt.GenerateFromPassword(salted, passwordHashCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *userStorage) checkPassword(password, hash string) bool {
	salted := u.saltedPassword(password)
	return bcrypt.CompareHashAndPassword([]byte(hash), salted) == nil
}

func (u *userStorage) saltedPassword(password string) []byte {
	b := []byte(u.secret)
	salted := append(b, []byte(password)...)
	return salted
}

func (u *userStorage) randomToken() string {
	b := make([]rune, tokenLength)
	for i := range b {
		b[i] = tokenRunes[rand.Intn(len(tokenRunes))]
	}
	return string(b)
}
