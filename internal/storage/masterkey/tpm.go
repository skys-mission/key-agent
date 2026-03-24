// Package masterkey provides the TPM 2.0 backend.
package masterkey

import (
	"os"
	"runtime"
)

// TPMBackend stores the master key using TPM 2.0 hardware.
type TPMBackend struct {
	devicePath string
}

// NewTPMBackend creates a new TPM backend.
func NewTPMBackend() *TPMBackend {
	devicePath := "/dev/tpm0"
	if runtime.GOOS == "windows" {
		devicePath = "" // Windows uses TBS (TPM Base Services)
	}

	return &TPMBackend{
		devicePath: devicePath,
	}
}

// Name returns the backend name.
func (b *TPMBackend) Name() string {
	return "tpm"
}

// Available returns true if TPM is accessible.
func (b *TPMBackend) Available() bool {
	if runtime.GOOS == "windows" {
		// Windows has TPM via TBS, need to check differently
		// For now, we'll check if we can access TPM
		// This requires actual TPM communication
		return false // TODO: Implement Windows TPM check
	}

	// Linux: check if TPM device exists
	if _, err := os.Stat(b.devicePath); err != nil {
		return false
	}

	// TODO: Actually verify TPM is functional by opening a session
	// For MVP, we'll just check if the device file exists
	return true
}

// Get retrieves the master key from TPM.
func (b *TPMBackend) Get() ([]byte, error) {
	// TODO: Implement TPM key retrieval
	// This requires:
	// 1. Opening TPM device
	// 2. Loading the key from TPM's NVRAM or sealed storage
	// 3. Unsealing the key (requires PCR state match)
	//
	// For MVP, we return not found to fall back to other backends
	return nil, ErrKeyNotFound
}

// Set stores the master key in TPM.
func (b *TPMBackend) Set(key []byte) error {
	// TODO: Implement TPM key storage
	// This requires:
	// 1. Opening TPM device
	// 2. Sealing the key to TPM NVRAM with PCR policy
	// 3. Storing the sealed blob
	//
	// For MVP, we return error to fall back to other backends
	return ErrNotAvailable
}

// Delete removes the master key from TPM.
func (b *TPMBackend) Delete() error {
	// TODO: Implement TPM key deletion
	return nil
}
