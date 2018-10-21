package solidity

import (
	"errors"
)

// solidity errors
var (
	ErrExistAddress        = errors.New("exist address")
	ErrInvalidSequence     = errors.New("invalid sequence")
	ErrInsuffcientBalance  = errors.New("insufficient balance")
	ErrVirtualMachinePanic = errors.New("virtual machine panic")
	ErrNotAllowed          = errors.New("not allowed")
)
