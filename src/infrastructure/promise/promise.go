package promise

import "fmt"

//422 - Validation Errors, an array of objects, each object containing the field and the value (message) of the error
//400 - Bad Request - Any request not properly formatted for the server to understand and parse it
//403 - Forbidden - This can be used for any authentication errors, a user not being logged in etc.
//404 - Any requested entity which is not being found on the server
//406 - Not Accepted - The example usage for this code, is an attempt on an expired or timed-out action. Such as trying to cancel an order which cannot be cancelled any more
//409 - Conflict - Anything which causes conflicts on the server, the most famous one, a not unique email error, a duplicate entity...

const (
	BadRequest			= 400
	ForBidden			= 403
	NotFound			= 404
	NotAccepted			= 406
	Conflict			= 409
	ValidationError 	= 422
	InternalError		= 500
)


type DataChan <-chan FutureData

type IPromise interface {
	GetData() 	DataChan
	Count()		int
	Capacity()	int
}

type FutureData struct {
	Data 	interface{}
	Error 	error
}

type FutureError struct {
	Code 	int32
	Reason	string
}

func (error FutureError) Error() string {
	return fmt.Sprintf("err code: %d, reason: %s", error.Code, error.Reason)
}