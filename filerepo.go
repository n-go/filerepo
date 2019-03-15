package filerepo

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

// Repo represents a file content mananget
type Repo struct {
	lock      sync.RWMutex
	location  string
	hashFunc  func(data []byte) uint64
	hashValue map[string]uint64
}

// New returns a new instance of management.
func New(location string, recursive bool, extensions []string, hashFunc func(data []byte) uint64) *Repo {
	if hashFunc == nil {
		panic(errors.New("Hash function is not provided for filerepo"))
	}

	r := &Repo{
		sync.RWMutex{},
		location,
		hashFunc,
		map[string]uint64{},
	}

	// Create directory if not existing.
	if _, err := os.Stat(location); os.IsNotExist(err) {
		err = os.MkdirAll(location, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	r.scan("", recursive, extensions)
	return r
}

// Hash returns the hash of the file content, 0 if the file is not available.
func (r *Repo) Hash(path string) uint64 {
	r.lock.RLock()
	value, ok := r.hashValue[path]
	r.lock.RUnlock()

	if !ok {
		return 0
	}
	if value != 0 {
		return value
	}

	data, err := ioutil.ReadFile(filepath.Join(r.location, path))
	if err != nil {
		return 0
	}
	value = r.hashFunc(data)

	r.lock.Lock()
	r.hashValue[path] = value
	r.lock.Unlock()
	return value
}

// Read returns the file content if the file is not available.
func (r *Repo) Read(filename string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(r.location, filename))
}

// Save saves the file content.
func (r *Repo) Save(path string, data []byte) error {
	if err := ioutil.WriteFile(filepath.Join(r.location, path), data, os.ModePerm); err != nil {
		return err
	}
	value := r.hashFunc(data)

	r.lock.Lock()
	r.hashValue[path] = value
	r.lock.Unlock()

	return nil
}

func (r *Repo) scan(dir string, recursive bool, extensions []string) {
	files, err := ioutil.ReadDir(filepath.Join(r.location, dir))
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		name := file.Name()
		if dir != "" {
			name = dir + "/" + name
		}

		if file.IsDir() {
			if recursive {
				r.scan(name, recursive, extensions)
			}
			continue
		}

		if extensions != nil {
			ext := filepath.Ext(file.Name())
			matched := false
			for _, e := range extensions {
				if e == ext {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		r.hashValue[name] = 0
	}
}
