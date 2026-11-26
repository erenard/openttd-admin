// GSException mirrors com.openttd.gamescript.GSException: the RuntimeException
// GSRequest.getResult() throws when AdminCmd answered with an "exception" field
// instead of "data".

package gamescript

type GSException struct {
	message string
}

// NewException is GSException(String exception).
func NewException(message string) *GSException {
	return &GSException{message: message}
}

// Error is getMessage, and makes the error interface work.
func (e *GSException) Error() string { return e.message }

// GetExceptionMessage is getMessage, for callers that do not want the error
// interface.
func (e *GSException) GetExceptionMessage() string { return e.message }

// ToString is Java's Throwable.toString(): the class name and the message.
func (e *GSException) ToString() string {
	return "GSException: " + e.message
}
