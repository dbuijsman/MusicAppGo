package common

// DBError will be the type for returning errors that arise from requests to a database
type DBError struct {
	ErrorCode    int
	ErrorMessage string
}

func (err DBError) Error() string {
	return err.ErrorMessage
}

// DuplicateEntry is errorcode for duplicate entries
const DuplicateEntry int = 2

// UnknownError is errorcode for unspecified errors
const UnknownError int = 500

// InvalidOffsetMax is errorcode for requests with an invalid offset or invalid max
const InvalidOffsetMax int = 400

// ScannerError is errorcode for requests where there went something wrong with scanning the rows
const ScannerError int = 510

// GetDBError for returning errors with errormessages
func GetDBError(message string, code int) DBError {
	return DBError{ErrorCode: code, ErrorMessage: message}
}
