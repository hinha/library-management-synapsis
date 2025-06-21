package seeder

import (
	"context"
	"errors"
	"github.com/hinha/library-management-synapsis/cmd/config"
	"github.com/hinha/library-management-synapsis/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
	"log"
)

// UserSeeder handles database seeding
type UserSeeder struct {
	userRepo user.IRepository
}

// NewUserSeeder creates a new seeder
func NewUserSeeder(userRepo user.IRepository) *UserSeeder {
	return &UserSeeder{
		userRepo: userRepo,
	}
}

// SeedUsers seeds initial users into the database
func (s *UserSeeder) SeedUsers(ctx context.Context) error {
	// Check if admin user already exists
	existingAdmin, err := s.userRepo.GetByEmail(ctx, config.InitialAdminEmail)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return err
	}

	// If admin user doesn't exist, create it
	if existingAdmin == nil {
		// Create admin user with bcrypt hashed password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(config.InitialAdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		adminUser := &user.User{
			Name:     "Admin",
			Email:    config.InitialAdminEmail,
			Password: string(hashedPassword),
			Role:     user.RoleAdmin,
		}

		if err := s.userRepo.Create(ctx, adminUser); err != nil {
			return err
		}

		// Verify that the user was created correctly by retrieving it from the database
		createdUser, err := s.userRepo.GetByEmail(ctx, config.InitialAdminEmail)
		if err != nil {
			return err
		}

		if createdUser == nil {
			return errors.New("failed to retrieve created admin user")
		}

		// Verify that the password is correct
		if !createdUser.ComparePassword(config.InitialAdminPassword) {
			return errors.New("password verification failed for admin user")
		}

		log.Println("Admin user created and verified successfully")
	}

	return nil
}

// Seed seeds all initial data
func (s *UserSeeder) Seed(ctx context.Context) error {
	if err := s.SeedUsers(ctx); err != nil {
		return err
	}

	return nil
}
