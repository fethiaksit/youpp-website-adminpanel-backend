package handlers

import "errors"

var (
	errMissingUserContext = errors.New("missing user context")
	errInvalidUserContext = errors.New("invalid user context")
	errMissingRoleContext = errors.New("missing role context")
)
