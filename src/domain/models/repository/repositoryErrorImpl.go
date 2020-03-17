package repository

type iRepoError struct {
	ErrCode   ErrorCode
	ErrMsg    string
	ErrReason error
}

func ErrorFactory(code ErrorCode, message string, reason error) IRepoError {
	return &iRepoError{
		ErrCode:   code,
		ErrMsg:    message,
		ErrReason: reason,
	}
}

func (iRepo *iRepoError) Code() ErrorCode {
	return iRepo.ErrCode
}

func (iRepo *iRepoError) Message() string {
	return iRepo.ErrMsg
}

func (iRepo *iRepoError) Reason() error {
	return iRepo.ErrReason
}

func (iRepo *iRepoError) Error() string {
	return iRepo.ErrReason.Error()
}
