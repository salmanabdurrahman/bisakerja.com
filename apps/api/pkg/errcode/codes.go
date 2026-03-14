package errcode

const (
	BadRequest             = "BAD_REQUEST"
	Unauthorized           = "UNAUTHORIZED"
	Forbidden              = "FORBIDDEN"
	TooManyRequests        = "TOO_MANY_REQUESTS"
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
	InvalidPlanCode        = "INVALID_PLAN_CODE"
	InvalidRedirectURL     = "INVALID_REDIRECT_URL"
	InvalidWebhookPayload  = "INVALID_WEBHOOK_PAYLOAD"
	InvalidWebhookToken    = "INVALID_WEBHOOK_TOKEN" // #nosec G101 -- API error code, not secret material.
	WebhookUserNotFound    = "WEBHOOK_USER_NOT_FOUND"
	AlreadyPremium         = "ALREADY_PREMIUM"
	MayarRateLimited       = "MAYAR_RATE_LIMITED"
	MayarUpstreamError     = "MAYAR_UPSTREAM_ERROR"
	NotFound               = "NOT_FOUND"
	ServiceUnavailable     = "SERVICE_UNAVAILABLE"
	InternalServerError    = "INTERNAL_SERVER_ERROR"
)
