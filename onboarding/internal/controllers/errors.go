package controllers

import (
	"errors"
)

var (
	errMissingAppId  = errors.New("appId query parameter is required")
	errMissingAppKey = errors.New("X-App-Key header is required")
)
