package room

import "backend/internal/apperr"

// All sentinels are *apperr.Error and carry the correct HTTP-mappable Code.
var (
	// ErrRoomNotFound is returned when a room code does not exist in the repository.
	ErrRoomNotFound error = apperr.New(apperr.CodeNotFound, "room not found")

	// ErrRoomAlreadyStarted is returned when attempting to join a room that has already started.
	// Carries CodeConflict so transporters map it to HTTP 409, not 403.
	ErrRoomAlreadyStarted error = apperr.New(apperr.CodeConflict, "room already started")

	// ErrNoPlacesFound is returned when the upstream places provider returned zero results.
	ErrNoPlacesFound error = apperr.New(apperr.CodeNotFound, "no places found within the specified area")

	// ErrInvalidFilters is returned when search filters fail domain validation.
	ErrInvalidFilters error = apperr.New(apperr.CodeInvalid, "invalid search filters")

	// ErrInvalidVote is returned when a vote payload is malformed.
	ErrInvalidVote error = apperr.New(apperr.CodeInvalid, "invalid vote payload")
)
