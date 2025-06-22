package domain

import (
	"errors"
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
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string         `gorm:"size:255;not null" json:"-"`
	Role      Role           `gorm:"not null;default:'operation'" json:"role"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// NewUser creates a new user entity
func NewUser(name, email, password string, role Role) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("invalid user data")
	}
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
	var role pb.UserRole
	if u.Role == RoleAdmin {
		role = pb.UserRole_USER_ROLE_ADMIN
	} else if u.Role == RoleOperation {
		role = pb.UserRole_USER_ROLE_OPERATION
	} else {
		role = pb.UserRole_USER_ROLE_UNSPECIFIED
	}
	return &pb.UserResponse{
		Id:    strconv.Itoa(int(u.ID)),
		Name:  u.Name,
		Email: u.Email,
		Role:  role,
	}
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

func (u *User) UserIDString() string {
	return strconv.Itoa(int(u.ID))
}
