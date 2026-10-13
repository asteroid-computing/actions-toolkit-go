package core

import (
	"runtime"
	"uuid"
)

var OsSpecificNewline string

func init() {
	if runtime.GOOS == "windows" {
		OsSpecificNewline = "\r\n"
	} else {
		OsSpecificNewline = "\n"
	}
}

// lookupEnvFunc defines a function that returns the value of an environment
// variable and whether it was set.
type lookupEnvFunc func(string) (string, bool)

type uuidGeneratorFunc func() uuid.UUID

// setenvFunc sets an environment variable in the current process, like os.Setenv.
type setenvFunc func(key, value string) error
