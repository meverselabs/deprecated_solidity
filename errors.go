package solidity

import (
	"errors"
)

// solidity errors
var (
	ErrExistAddress        = errors.New("exist address")
	ErrExistAccountName    = errors.New("exist account name")
	ErrInvalidAccountName  = errors.New("invaild account name")
	ErrInvalidSequence     = errors.New("invalid sequence")
	ErrInsuffcientBalance  = errors.New("insufficient balance")
	ErrVirtualMachinePanic = errors.New("virtual machine panic")
	ErrInvalidSignerCount  = errors.New("invalid signer count")
	ErrNotAllowed          = errors.New("not allowed")
)
