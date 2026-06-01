package mach

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"

	"github.com/machbase/neo-engine/v8/native"
)

func LinkVersion() string {
	return native.Version
}

func LinkGitHash() string {
	return native.GitHash
}

type InitOption int

const (
	// machbase-engine takes all control of the signals
	OPT_SIGHANDLER_ON InitOption = 0x0
	// the caller takes all control, machbase-engine can not leave stack dump when the process crashed
	OPT_SIGHANDLER_OFF InitOption = 0x1
	// engine takes all control except SIGINT, so that the caller can take SIGINT control
	OPT_SIGHANDLER_SIGINT_OFF InitOption = 0x2
)

type Env struct {
	sync.Mutex
	handle    unsafe.Pointer
	onceStart sync.Once
	onceStop  sync.Once
}

var _env = Env{}

var ErrDatabaseNotInitialized = errors.New("database not initialized")

func Initialize(homeDir string, machPort int, opt InitOption) error {
	_env.Lock()
	defer _env.Unlock()
	if _env.handle != nil {
		return fmt.Errorf("database already initialized")
	}
	homeDir = translateCodePage(homeDir)
	var handle unsafe.Pointer
	err := EngInitialize(homeDir, machPort, int(opt), &handle)
	if err != nil {
		return err
	}
	_env.handle = handle
	return nil
}

func Finalize() {
	_env.Lock()
	defer _env.Unlock()
	if _env.handle != nil {
		EngFinalize(_env.handle)
	}
}

func DestroyDatabase() error {
	_env.Lock()
	defer _env.Unlock()
	if _env.handle == nil {
		return ErrDatabaseNotInitialized
	}
	return EngDestroyDatabase(_env.handle)
}

func CreateDatabase() error {
	_env.Lock()
	defer _env.Unlock()
	if _env.handle == nil {
		return ErrDatabaseNotInitialized
	}
	return EngCreateDatabase(_env.handle)
}

func ExistsDatabase() bool {
	_env.Lock()
	defer _env.Unlock()
	if _env.handle == nil {
		return false
	}
	return EngExistsDatabase(_env.handle)
}

func RestoreDatabase(path string) error {
	return EngRestoreDatabase(_env.handle, path)
}

func StartDatabase() (err error) {
	_env.Lock()
	defer _env.Unlock()
	_env.onceStart.Do(func() {
		err = EngStartup(_env.handle)
	})
	return
}

func StopDatabase() (err error) {
	_env.Lock()
	defer _env.Unlock()
	_env.onceStop.Do(func() {
		if _env.handle != nil {
			err = EngShutdown(_env.handle)
			_env.handle = nil
		}
	})
	return
}

func ErrDatabase() error {
	return EngError(_env.handle)
}
