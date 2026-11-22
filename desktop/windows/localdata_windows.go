//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

type cryptBlob struct {
	size uint32
	data *byte
}

var (
	crypt32        = syscall.NewLazyDLL("crypt32.dll")
	cryptProtect   = crypt32.NewProc("CryptProtectData")
	cryptUnprotect = crypt32.NewProc("CryptUnprotectData")
	cryptLocalFree = syscall.NewLazyDLL("kernel32.dll").NewProc("LocalFree")
	moveFileEx     = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")
)

const cryptProtectUIForbidden = 0x1
const moveReplaceExisting = 0x1
const moveWriteThrough = 0x8

func cryptCall(source []byte, protect bool) ([]byte, error) {
	if len(source) == 0 {
		return nil, errors.New("empty private data")
	}
	in := cryptBlob{size: uint32(len(source)), data: &source[0]}
	var out cryptBlob
	call := cryptProtect
	if !protect {
		call = cryptUnprotect
	}
	ok, _, sysErr := call.Call(uintptr(unsafe.Pointer(&in)), 0, 0, 0, 0, cryptProtectUIForbidden, uintptr(unsafe.Pointer(&out)))
	if ok == 0 {
		return nil, fmt.Errorf("Windows data protection failed: %v", sysErr)
	}
	defer cryptLocalFree.Call(uintptr(unsafe.Pointer(out.data)))
	if out.data == nil || out.size == 0 {
		return nil, errors.New("Windows data protection returned no data")
	}
	return append([]byte(nil), unsafe.Slice(out.data, int(out.size))...), nil
}
func protectLocal(plain []byte) ([]byte, error)       { return cryptCall(plain, true) }
func unprotectLocal(encrypted []byte) ([]byte, error) { return cryptCall(encrypted, false) }

func atomicPrivateWrite(path string, encrypted []byte) error {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".encrypted-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(encrypted); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	from, err := syscall.UTF16PtrFromString(tmp)
	if err != nil {
		return err
	}
	to, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	ok, _, sysErr := moveFileEx.Call(uintptr(unsafe.Pointer(from)), uintptr(unsafe.Pointer(to)), moveReplaceExisting|moveWriteThrough)
	if ok == 0 {
		return fmt.Errorf("encrypted file replacement failed: %v", sysErr)
	}
	return nil
}
