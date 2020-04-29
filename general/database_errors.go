package general

// DBError will be the type for returning errors that arise from requests to a database
type DBError struct {
	ErrorCode    int
	ErrorMessage string
}

func (err DBError) Error() string {
	return err.ErrorMessage
}

// InvalidOffsetMax is errorcode for requests with an invalid offset or invalid max
const InvalidOffsetMax int = 400

// NotFoundError is errorcode if the requested entry can not be found
const NotFoundError int = 404

// InvalidInput is errorcode for when the given values do not satisfy the requirements in the database
const InvalidInput int = 406

// DuplicateEntry is errorcode for duplicate entries
const DuplicateEntry int = 409

// MissingForeignKey is errorcode for when inserting a value for a foreign key that will point to a non-existing row
const MissingForeignKey int = 412

// UnknownError is errorcode for unspecified errors
const UnknownError int = 500

// ScannerError is errorcode for requests where there went something wrong with scanning the rows
const ScannerError int = 510

// GetDBError for returning errors with errormessages
func GetDBError(message string, code int) DBError {
	return DBError{ErrorCode: code, ErrorMessage: message}
}

// ErrorToUnknownDBError converts the error to a DBError with UnknownError as ErrorCode
func ErrorToUnknownDBError(err error) DBError {
	return DBError{ErrorCode: UnknownError, ErrorMessage: err.Error()}
}

// MySQLErrorToDBError translates an error from a query to MySQL into a DBError
func MySQLErrorToDBError(err error) DBError {
	switch err.Error()[6:10] {
	case "1062":
		return GetDBError(err.Error(), DuplicateEntry)
	case "1452":
		return GetDBError(err.Error(), MissingForeignKey)
	default:
		return ErrorToUnknownDBError(err)
	}
}
