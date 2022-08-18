package service

import (
	"context"

	"github.com/Edbeer/Project/pkg/utils"

	"github.com/Edbeer/Project/internal/entity"

	"github.com/Edbeer/Project/config"
)

// Token Manager interface
type Manager interface {
	GenerateJWTToken(user *entity.User) (string, error)
	Parse(accessToken string) (string, error)
	NewRefreshToken() string
}

// User psql storage interface
type UserPsql interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	FindUserByEmail(ctx context.Context, user *entity.User) (*entity.User, error)
}

// PasswordHasher provides hashing logic to securely store passwords
type PasswordHasher interface {
	Hash(password string) string
}


// User service
type UserService struct {
	config *config.Config
	psql   UserPsql
	hash PasswordHasher
	tokenManager Manager
}

// New user service constructor
func NewUserService(config *config.Config, psql UserPsql, hash PasswordHasher, tokenManager Manager) *UserService {
	return &UserService{
		config: config,
		psql:   psql,
		hash: hash,
		tokenManager: tokenManager,
	}
}

func (u *UserService) SignUp(ctx context.Context, input *entity.InputUser) (*entity.UserWithToken, error) {	
	user := &entity.User{
		Name: input.Name,
		Password: u.hash.Hash(input.Password),
		Email: input.Email,
	}

	if err := user.PrepareCreate(); err != nil {
		return nil, err
	}

	existsUser, err := u.psql.FindUserByEmail(ctx, user)
	if existsUser != nil || err == nil {
		return nil, err
	}

	if err := utils.ValidateStruct(ctx, user); err != nil {
		return nil, err
	}

	createdUser, err := u.psql.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	accessToken, err := u.tokenManager.GenerateJWTToken(createdUser)
	if err != nil {
		return nil, err
	}
	
	return &entity.UserWithToken{
		User:  createdUser,
		AccessToken: accessToken,
	}, nil
}

func (u *UserService) SignIn(ctx context.Context, user *entity.User) (*entity.UserWithToken, error) {
	foundUser, err := u.psql.FindUserByEmail(ctx, user)
	if err != nil {
		return nil, err
	}

	accessToken, err := u.tokenManager.GenerateJWTToken(foundUser)
	if err != nil {
		return nil, err
	}
	
	return &entity.UserWithToken{
		User:  foundUser,
		AccessToken: accessToken,
	}, nil
}