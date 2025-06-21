package validator

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/hinha/library-management-synapsis/gen/api/proto/common"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

var validate = validator.New()

var (
	InvalidParam = errors.New("invalid parameter")
)

func init() {
	_ = RegisterEnumValidator(validate)
}

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		return NewValidationGRPCError(err)
	}
	return nil
}

// NewValidationGRPCError creates a gRPC error with embedded validation detail(s)
func NewValidationGRPCError(err error) error {
	var validateErrs validator.ValidationErrors
	ok := errors.As(err, &validateErrs)
	if !ok {
		// If not a validation error, return as internal error
		return gstatus.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	var detailList []*anypb.Any

	for _, fieldErr := range validateErrs {
		detail := &common.FieldValidationError{
			Field:   fieldErr.Field(),
			Message: fmt.Sprintf("failed on '%s' validation", fieldErr.ActualTag()),
		}

		anyTo, err := anypb.New(detail)
		if err != nil {
			// fallback if marshal to Any fails
			continue
		}

		detailList = append(detailList, anyTo)
	}

	statusProto := &spb.Status{
		Code:    int32(codes.InvalidArgument),
		Message: InvalidParam.Error(),
		Details: detailList,
	}

	return gstatus.FromProto(statusProto).Err()
}
