package errcode

const (
	BadRequest             = "BAD_REQUEST"
	Unauthorized           = "UNAUTHORIZED"
	Forbidden              = "FORBIDDEN"
	EmailAlreadyRegistered = "EMAIL_ALREADY_REGISTERED"
	InvalidCredentials     = "INVALID_CREDENTIALS" // #nosec G101 -- API error code, not secret material.
	InvalidEmail           = "INVALID_EMAIL"
	InvalidPassword        = "INVALID_PASSWORD"
	InvalidName            = "INVALID_NAME"
	InvalidLimit           = "INVALID_LIMIT"
	InvalidSort            = "INVALID_SORT"
	InvalidSource          = "INVALID_SOURCE"
	InvalidPage            = "INVALID_PAGE"
	InvalidSalaryMin       = "INVALID_SALARY_MIN"
	InvalidJobType         = "INVALID_JOB_TYPE"
	NotFound               = "NOT_FOUND"
	InternalServerError    = "INTERNAL_SERVER_ERROR"
)
