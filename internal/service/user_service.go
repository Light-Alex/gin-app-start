package service

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"gin-app-start/internal/common"
	"gin-app-start/internal/dto"
	"gin-app-start/internal/model"
	"gin-app-start/internal/repository"
	"gin-app-start/pkg/errors"

	"gorm.io/gorm"
)

type UserService interface {
	Login(ctx common.Context, req *dto.LoginRequest) (*model.User, error)
	CreateUser(ctx common.Context, req *dto.CreateUserRequest) (*model.User, error)
	UpdatePassword(ctx common.Context, req *dto.UpdatePasswordRequest) error
	UploadImage(ctx common.Context, username, filename string) error
	GetUser(ctx common.Context, id uint) (*model.User, error)
	GetUserByUsername(ctx common.Context, username string) (*model.User, error)
	UpdateUser(ctx common.Context, id uint, req *dto.UpdateUserRequest) (*model.User, error)
	DeleteUser(ctx common.Context, id uint) error
	ListUsers(ctx common.Context, page, pageSize int) ([]*model.User, int64, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) CreateUser(ctx common.Context, req *dto.CreateUserRequest) (*model.User, error) {
	existingUser, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("User already exists")
	}

	if req.Email != "" {
		existingUser, err = s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.New("Email already exists")
		}
	}

	salt := generateSalt()
	hashedPassword := hashPassword(req.Password, salt)

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: hashedPassword,
		Salt:     salt,
		Status:   1,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(ctx common.Context, req *dto.LoginRequest) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	if !VerifyPassword(req.Password, user.Salt, user.Password) {
		return nil, errors.New("Password not match")
	}

	return user, nil
}

func (s *userService) UpdatePassword(ctx common.Context, req *dto.UpdatePasswordRequest) error {
	user, err := s.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return err
	}

	// 验证旧密码是否匹配
	if !VerifyPassword(req.OldPassword, user.Salt, user.Password) {
		return errors.New("Old password error")
	}

	// 生成新的盐值和哈希密码
	newSalt := generateSalt()
	newHashedPassword := hashPassword(req.NewPassword, newSalt)

	// 更新用户密码和盐值
	user.Salt = newSalt
	user.Password = newHashedPassword

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *userService) UploadImage(ctx common.Context, username, filename string) error {
	user, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return err
	}

	user.Avatar = filename
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *userService) GetUser(ctx common.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) GetUserByUsername(ctx common.Context, username string) (*model.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx common.Context, id uint, req *dto.UpdateUserRequest) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != 0 {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(ctx common.Context, id uint) error {
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *userService) ListUsers(ctx common.Context, page, pageSize int) ([]*model.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize
	users, total, err := s.userRepo.List(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func generateSalt() string {
	salt := make([]byte, 16)
	rand.Read(salt)
	return hex.EncodeToString(salt)
}

func hashPassword(password, salt string) string {
	hash := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

func VerifyPassword(password, salt, hashedPassword string) bool {
	return hashPassword(password, salt) == hashedPassword
}
