// Package config provides configuration for the application.
package config

import "errors"

// ErrPublicKeyPathNotProvided is an error indicating that a public key path was not provided.
var (
	ErrPublicKeyPathNotProvided = errors.New("public key path not provided") // Error for missing public key path.
	ErrPrivateKeyPathNotSet     = errors.New("private key path not set")     // Error for missing private key path.
)
