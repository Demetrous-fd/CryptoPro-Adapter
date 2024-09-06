package cades

import "errors"

var (
	ErrUnknownCallback = errors.New("unknown callback")
	ErrMethodExecution = errors.New("the method cannot be executed")
	ErrProperty        = errors.New("fail to get/set property")
	ErrEmpty           = errors.New("empty")
	ErrContainerExists = errors.New("container exists")
)
