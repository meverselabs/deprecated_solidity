package vm

import (
	"math/big"
)

const (
	MaximumExtraDataSize uint64 = 32   // Maximum size extra data may be after Genesis.
	QuadCoeffDiv         uint64 = 512  // Divisor for the quadratic particle of the memory cost equation.
	CallStipend          uint64 = 2300 // Free given at beginning of call.

	EpochDuration   uint64 = 30000 // Duration between proof-of-work epochs.
	CallCreateDepth uint64 = 1024  // Maximum depth of call/create stack.
	StackLimit      uint64 = 1024  // Maximum size of VM stack allowed.

	MaxCodeSize = 24576 // Maximum bytecode to permit for a contract

	ModExpQuadCoeffDiv uint64 = 20 // Divisor for the quadratic particle of the big int modular exponentiation
)

var (
	DifficultyBoundDivisor = big.NewInt(2048)   // The bound divisor of the difficulty, used in the update calculations.
	GenesisDifficulty      = big.NewInt(131072) // Difficulty of the Genesis block.
	MinimumDifficulty      = big.NewInt(131072) // The minimum that the difficulty may ever be.
	DurationLimit          = big.NewInt(13)     // The decision boundary on the blocktime duration used to determine whether difficulty should go up or not.
)
