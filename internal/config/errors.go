package config

import "errors"

var (
	ErrPublicKeyPathNotProvided = errors.New("public key path not provided") // Error for missing public key path.
	ErrPrivateKeyPathNotSet     = errors.New("private key path not set")     // Error for missing private key path.
)
