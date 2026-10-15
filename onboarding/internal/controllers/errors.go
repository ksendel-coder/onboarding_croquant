package controllers

import (
	"errors"
)

var (
	errMissingAppId = errors.New("appId query parameter is required")
)
