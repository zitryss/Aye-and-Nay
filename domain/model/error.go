package model

import (
	"errors"
)

var (
	ErrTooManyRequests       = errors.New("too many requests")
	ErrBodyTooLarge          = errors.New("body too large")
	ErrWrongContentType      = errors.New("wrong content type")
	ErrNotEnoughImages       = errors.New("not enough images")
	ErrTooManyImages         = errors.New("too many images")
	ErrImageTooLarge         = errors.New("image too large")
	ErrNotImage              = errors.New("not image")
	ErrDurationNotSet        = errors.New("duration not set")
	ErrDurationInvalid       = errors.New("duration invalid")
	ErrAlbumNotFound         = errors.New("album not found")
	ErrPairNotFound          = errors.New("pair not found")
	ErrTokenNotFound         = errors.New("token not found")
	ErrImageNotFound         = errors.New("image not found")
	ErrAlbumAlreadyExists    = errors.New("album already exists")
	ErrTokenAlreadyExists    = errors.New("token already exists")
	ErrUnsupportedMediaType  = errors.New("unsupported media type")
	ErrThirdPartyUnavailable = errors.New("third party unavailable")
	ErrUnknown               = errors.New("unknown")
)
