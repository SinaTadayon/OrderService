package repository

type iRepoError struct {
	code    ErrorCode
	message string
	reason  error
}

func ErrorFactory(code ErrorCode, message string, reason error) IRepoError {
	return &iRepoError{
		code:    code,
		message: message,
		reason:  reason,
	}
}

func (iRepo *iRepoError) Code() ErrorCode {
	return iRepo.code
}

func (iRepo *iRepoError) Message() string {
	return iRepo.message
}

func (iRepo *iRepoError) Reason() error {
	return iRepo.reason
}

func (iRepo *iRepoError) Error() string {
	return iRepo.reason.Error()
}
