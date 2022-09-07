package metrics

const (
	instrumentationName  = "keyvaultkms"
	errorMessageKey      = "error_message"
	statusTypeKey        = "status"
	operationTypeKey     = "operation"
	kmsRequestMetricName = "kms_request"
	// ErrorStatusTypeValue sets status tag to "error"
	ErrorStatusTypeValue = "error"
	// SuccessStatusTypeValue sets status tag to "success"
	SuccessStatusTypeValue = "success"
	// EncryptOperationTypeValue sets operation tag to "encrypt"
	EncryptOperationTypeValue = "encrypt"
	// DecryptOperationTypeValue sets operation tag to "decrypt"
	DecryptOperationTypeValue = "decrypt"
	// GrpcOperationTypeValue sets operation tag to "grpc"
	GrpcOperationTypeValue = "grpc"
)
