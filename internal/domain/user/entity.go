package user

import (
	"strconv"
	"time"

	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Role represents user roles in the system
type Role string

const (
	// RoleAdmin represents an admin user with elevated privileges
	RoleAdmin Role = "admin"
	// RoleOperation represents a operation user with standard privileges
	RoleOperation Role = "operation"
)

// User represents a user entity in the system
type User struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	Email     string         `gorm:"uniqueIndex;size:255;not null"`
	Password  string         `gorm:"size:255;not null"`
	Role      Role           `gorm:"not null;default:'operation'"`
	Active    bool           `gorm:"default:true"`
	CreatedAt time.Time      `gorm:"not null"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// NewUser creates a new user entity
func NewUser(name, email, password string, role Role) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

// ComparePassword compares the provided password with the user's hashed password
func (u *User) ComparePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// ToProto converts the user entity to a protobuf user response
func (u *User) ToProto() *pb.UserResponse {
	return &pb.UserResponse{
		Id:    strconv.Itoa(int(u.ID)),
		Name:  u.Name,
		Email: u.Email,
	}
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}
