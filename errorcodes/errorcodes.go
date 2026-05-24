package errorcodes

// Service-level error codes returned in gRPC status messages.
const (
	ProfileNotFound          = "PROFILE_NOT_FOUND"
	ProfileAlreadyRegistered = "PROFILE_ALREADY_REGISTERED"
	Internal                 = "INTERNAL"
	InvalidArgument          = "INVALID_ARGUMENT"
	AIProviderError          = "AI_PROVIDER_ERROR"
	InsufficientBalance      = "INSUFFICIENT_BALANCE"
)
