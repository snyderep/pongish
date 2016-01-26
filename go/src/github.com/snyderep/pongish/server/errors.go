package server

import (
	"errors"
	"fmt"
)

// ErrMisingTemplateHandler is returned when a template exists in the templates directory but
// no template handler is specified.
type ErrMisingTemplateHandler struct {
	TemplateName string
}

func (e ErrMisingTemplateHandler) Error() string {
	return fmt.Sprintf("template %s is missing a template handler", e.TemplateName)
}

// ErrTooManyWaiting is returned when there are already too many waiting to play.
var ErrTooManyWaiting = errors.New("server: too many waiting")

// ErrMustBeWaitingState is returned when a player is required to be in waiting state but is not.
var ErrMustBeWaitingState = errors.New("server: player must be in waiting state")
