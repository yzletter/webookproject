package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"webook/internal/domain"
	"webook/internal/repository"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvaildUserOrPassword = errors.New("账号或密码不对")
)

type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, u domain.User) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error
	FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error)
	GetProfile(ctx context.Context, id int64) (domain.User, error)
}

type MyUserService struct {
	repo *repository.CachedUserRepository
}

func NewUserService(repo *repository.CachedUserRepository) *MyUserService {
	return &MyUserService{
		repo: repo,
	}
}

func (svc *MyUserService) SignUp(ctx context.Context, u domain.User) error {
	// 考虑加密放在哪里的问题
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是存起来
	return svc.repo.Create(ctx, u)
}

func (svc *MyUserService) Login(ctx context.Context, u domain.User) (domain.User, error) {
	// 先找用户
	r, err := svc.repo.FindByEmail(ctx, u.Email)
	if err == repository.ErrUserNotFound {
		return domain.User{}, ErrInvaildUserOrPassword
	}

	// 比较密码
	err = bcrypt.CompareHashAndPassword([]byte(r.Password), []byte(u.Password))
	if err != nil {
		// DEBUG
		return domain.User{}, ErrInvaildUserOrPassword
	}
	return r, err
}

func (svc *MyUserService) UpdateNonSensitiveInfo(ctx context.Context, u domain.User) error {
	return svc.repo.UpdateNonZeroFields(ctx, u)
}

func (svc *MyUserService) FindOrCreateByPhone(ctx context.Context, phone string) (domain.User, error) {
	u, err := svc.repo.FindByPhone(ctx, phone)
	if err != repository.ErrUserNotFound {
		return u, err
	}
	u = domain.User{
		Phone: phone,
	}
	err = svc.repo.Create(ctx, u)
	if err != nil {
		return u, err
	}
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *MyUserService) GetProfile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}
