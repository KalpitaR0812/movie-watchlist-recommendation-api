package services

import (
	"errors"
	"movie-watchlist/internal/models"
	"movie-watchlist/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Register(username, email, password string) (*models.User, error) {
	// Check if email already exists
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, err // Real database error
	}
	if user != nil {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	user, err = s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, err // Real database error
	}
	if user != nil {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user = &models.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (*models.User, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *UserService) GetByID(id primitive.ObjectID) (*models.User, error) {
	return s.userRepo.FindByID(id)
}
