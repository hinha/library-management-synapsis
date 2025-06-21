package validator

import (
	"github.com/go-playground/validator/v10"
	pb "github.com/hinha/library-management-synapsis/gen/api/proto/user"
)

func RegisterEnumValidator(v *validator.Validate) error {
	return v.RegisterValidation("role", func(fl validator.FieldLevel) bool {
		val, ok := fl.Field().Interface().(pb.UserRole)
		if !ok {
			return false
		}

		switch val {
		case pb.UserRole_USER_ROLE_ADMIN, pb.UserRole_USER_ROLE_OPERATION:
			return true
		default:
			return false
		}
	})
}
