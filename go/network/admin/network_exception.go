// NetworkException mirrors com.openttd.network.admin.NetworkException: a
// RuntimeException carrying either a message or the name of a NetworkErrorCode.
// It implements the builtin error interface, so core.Message() reads it and
// NetworkAdmin's receive loop can treat a broken packet and a refused admin
// connection the same way Java's catch (IOException | NetworkException) does.

package admin

import "com.openttd/network/constant"

type NetworkException struct {
	message string
}

// NewException is NetworkException(String message).
func NewException(message string) *NetworkException {
	return &NetworkException{message: message}
}

// NewExceptionFromError is NetworkException(int errorId): Java loops over every
// NetworkErrorCode and keeps the toString() of the last one whose ordinal
// matches, which is the constant's name. An id outside the enum leaves Java's
// message null; here it is the empty string.
func NewExceptionFromError(errorId int) *NetworkException {
	var message = ""
	var codes = constant.NetworkErrorCodeValues()
	for i := 0; i < len(codes); i++ {
		if int(codes[i]) == errorId {
			message = constant.NetworkErrorCodeName(constant.NetworkErrorCode(errorId))
		}
	}
	return &NetworkException{message: message}
}

// Error is getMessage.
func (e *NetworkException) Error() string { return e.message }

// GetExceptionMessage is getMessage, for callers that do not want the error
// interface.
func (e *NetworkException) GetExceptionMessage() string { return e.message }

// ToString is Java's Throwable.toString(): the class name and the message.
func (e *NetworkException) ToString() string {
	return "NetworkException: " + e.message
}

// Describe is what NetworkAdmin and NetworkClient log for an error id:
// NetworkErrorCode.valueOf(errorId).toString(). Java would print the string
// "null" for an unknown id, because valueOf returns null and string
// concatenation accepts it; constant.NetworkErrorCodeName prints
// NETWORK_ERROR_NONE instead, which says more and never throws.
func DescribeError(errorId int) string {
	return constant.NetworkErrorCodeName(constant.NetworkErrorCode(errorId))
}
