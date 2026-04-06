package model

import "errors"

// Common errors
var (
	// Not found errors
	ErrUserNotFound          = errors.New("user not found")
	ErrSessionNotFound       = errors.New("session not found")
	ErrEventNotFound         = errors.New("event not found")
	ErrEventOptionNotFound   = errors.New("event option not found")
	ErrPaymentNotFound       = errors.New("payment not found")
	ErrPaymentRecordNotFound = errors.New("payment record not found")
	ErrVenueNotFound         = errors.New("venue not found")
	ErrParticipantNotFound   = errors.New("participant not found")
	ErrVoteNotFound          = errors.New("vote not found")

	// Validation errors
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidGoogleID = errors.New("invalid google id")
	ErrInvalidSession  = errors.New("invalid session")

	// Auth errors
	ErrUnauthorized   = errors.New("unauthorized")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidState   = errors.New("invalid state parameter")
	ErrInvalidCode    = errors.New("invalid authorization code")
	ErrSessionExpired = errors.New("session expired")

	// Business logic errors
	ErrNoParticipantsInEvent        = errors.New("cannot create payment: no participants in event")
	ErrAlreadyVoted                 = errors.New("already voted")
	ErrNotVoted                     = errors.New("not voted yet")
	ErrAlreadyJoined                = errors.New("already joined")
	ErrPaymentRecordNotConfirmed    = errors.New("payment record not confirmed")
	ErrPaymentConfigLocked          = errors.New("payment configuration can no longer be changed")
	ErrInviteOnlyRequiresLink       = errors.New("invite-only events can only be joined through the share link")
	ErrEventDeadlinePassed          = errors.New("event deadline has passed")
	ErrParticipantCapReached        = errors.New("participant cap has been reached")
	ErrSplitBillParticipantAssigned = errors.New("participant still has split bill items assigned")
	ErrSplitBillManualAdjustBlocked = errors.New("split bill payment records cannot be adjusted manually")
	ErrSplitBillChargeAllBlocked    = errors.New("split bill payments cannot use charge all")
	ErrSkipVotingRequiresOneOption  = errors.New("skip voting requires exactly one option")
	ErrEventScheduleEditNotAllowed  = errors.New("event schedule can only be edited while event is confirmed")
	ErrEventOptionEditNotAllowed    = errors.New("event options can only be edited while event is voting")

	// Event status errors
	ErrEventNotOpenForJoining = errors.New("event is not open for joining")
	ErrEventNotInVotingPhase  = errors.New("event is not in voting phase")
	ErrEventNotInPaymentPhase = errors.New("event is not in payment phase")
	ErrEventAlreadyCompleted  = errors.New("event is already completed")
	ErrEventCancelled         = errors.New("event is cancelled")

	// Database errors
	ErrDuplicateEmail    = errors.New("email already exists")
	ErrDuplicateGoogleID = errors.New("google account already linked")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err        error
	StatusCode int
	Message    string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(err error, statusCode int, message string) *AppError {
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
		Message:    message,
	}
}
