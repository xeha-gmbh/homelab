package api

// Service interface for a login attempt.
type Service interface {

	// Login with data in request.
	Login(request *Request) (*Response, error)
}
