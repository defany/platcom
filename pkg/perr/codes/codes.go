package codes

type Code uint

const (
	// OK is returned on success.
	OK Code = iota + 1

	// Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	//
	// The gRPC framework will generate this error code in the above two
	// mentioned cases.
	Unknown

	// InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	//
	// This error code will not be generated by the gRPC framework.
	InvalidArgument

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	//
	// This error code will not be generated by the gRPC framework.
	NotFound

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	//
	// This error code will not be generated by the gRPC framework.
	AlreadyExists

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
	//
	// This error code will not be generated by the gRPC core framework,
	// but expect authentication middleware to use it.
	PermissionDenied

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	//
	// This error code will be generated by the gRPC framework in
	// out-of-memory and server overload situations, or when a message is
	// larger than the configured maximum size.
	ResourceExhausted

	// Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	//
	// This error code will not be generated by the gRPC framework.
	Aborted

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	//
	// This error code will be generated by the gRPC framework. Most
	// commonly, you will see this error code when a method implementation
	// is missing on the server. It can also be generated for unknown
	// compression algorithms or a disagreement as to whether an RPC should
	// be streaming.
	Unimplemented

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	//
	// This error code will be generated by the gRPC framework in several
	// internal error conditions.
	Internal

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	//
	// This error code will be generated by the gRPC framework during
	// abrupt shutdown of a server process or network connection.
	Unavailable

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	//
	// The gRPC framework will generate this error code when the
	// authentication metadata is invalid or a Credentials callback fails,
	// but also expect authentication middleware to generate it.
	Unauthenticated
)
