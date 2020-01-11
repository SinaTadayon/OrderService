package repository

import "github.com/pkg/errors"

type ErrorCode int

const (
	BadRequestErr  ErrorCode = 400
	ForbiddenErr   ErrorCode = 403
	NotFoundErr    ErrorCode = 404
	NotAcceptedErr ErrorCode = 406
	ConflictErr    ErrorCode = 409
	ValidationErr  ErrorCode = 422
	InternalErr    ErrorCode = 500
)

var ErrorTotalCountExceeded = errors.New("total count exceeded")
var ErrorPageNotAvailable = errors.New("page not available")
var ErrorDeleteFailed = errors.New("update deletedAt field failed")
var ErrorRemoveFailed = errors.New("remove subpackage failed")
var ErrorUpdateFailed = errors.New("update subpackage failed")
var ErrorVersionUpdateFailed = errors.New("update subpackage version failed")

type IRepoError interface {
	error
	Code() ErrorCode
	Message() string
	Reason() error
}
