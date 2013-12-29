package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type mmapData struct {
	file  *os.File
	data  []byte
	count int
	sync.Mutex
}

func (md *mmapData) Decrement() bool {
	md.Lock()
	defer md.Unlock()

	md.count--
	if md.count == 0 {
		return true
	} else {
		return false
	}
}

type mmapCache struct {
	files map[string]*mmapData
	sync.Mutex
}

func NewMmapCache() *mmapCache {
	return &mmapCache{files: make(map[string]*mmapData)}
}

func (mc *mmapCache) MapFile(filename string, count int) (*mmapData, error) {
	mc.Lock()
	defer mc.Unlock()

	// File is already open and mmapped
	md, ok := mc.files[filename]
	if ok {
		return md, nil
	}

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	// Stat the file
	st, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := int(st.Size())

	// Mmap the file
	data, err := syscall.Mmap(int(file.Fd()), 0, (size+4095) & ^4095, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	// Cache the information and return the mmap
	mc.files[filename] = &mmapData{file: file, data: data, count: count}
	return mc.files[filename], nil
}

func (mc *mmapCache) CloseFile(filename string) error {
	mc.Lock()
	defer mc.Unlock()

	// Make sure the file is open first
	md, ok := mc.files[filename]
	if !ok {
		mc.Unlock()
		return fmt.Errorf("File '%s' is not open", filename)
	}

	// Make sure it has a 0 count
	if md.count > 0 {
		return fmt.Errorf("File '%s' does not have a 0 reference count", filename)
	}

	// Unmap the file
	err := syscall.Munmap(md.data)
	if err != nil {
		return err
	}

	// Close the file
	err = md.file.Close()
	if err != nil {
		return err
	}

	// Remove the file from the map
	delete(mc.files, filename)

	return nil
}
