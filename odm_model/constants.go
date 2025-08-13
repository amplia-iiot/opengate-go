package odm_model

const (
	Version      string = "1.0.0"
	RspOperation string = "resOperation"
	Description  string = "description"
	OpType       string = "operationType"
)

// known steps
type StepEnum string

const (
	ACCEPTED              StepEnum = "ACCEPTED"
	SET_DEVICE_PARAMETERS StepEnum = "SET_DEVICE_PARAMETERS"
	END_INSTALL           StepEnum = "ENDINSTALL"
)

type AsyncRspType string

const (
	SUCCESS             AsyncRspType = "SUCCESSFUL"
	OPERATION_PENDING   AsyncRspType = "OPERATION_PENDING"
	ERROR_IN_PARAM      AsyncRspType = "ERROR_IN_PARAM"
	NOT_SUPPORTED       AsyncRspType = "NOT_SUPPORTED"
	ALREADY_IN_PROGRESS AsyncRspType = "ALREADY_IN_PROGRESS"
	ERROR_PROCESSING    AsyncRspType = "ERROR_PROCESSING"
	ERROR_TIMEOUT       AsyncRspType = "ERROR_TIMEOUT"
	TIMEOUT_CANCELLED   AsyncRspType = "TIMEOUT_CANCELLED"
	CANCELLED           AsyncRspType = "CANCELLED"
	CANCELLED_INTERNAL  AsyncRspType = "CANCELLED_INTERNAL"
	ERROR               AsyncRspType = "ERROR"
	EMPTY_FOR_STEP      AsyncRspType = ""
	SKIPPED             AsyncRspType = "SKIPPED"
	NOT_EXECUTED        AsyncRspType = "NOT_EXECUTED"
)
const (
	NO_ERROR string = "No Error"
)

// searchEntities
const (
	FIELD_VALUE               string = "value"
	FIELD_ALIAS_IDENTIFICATOR string = "Identificator"
)
