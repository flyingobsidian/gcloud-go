package config

import "os"

// Env is the environment gcloud configuration is read from: environment
// variables (CLOUDSDK_CONFIG, CLOUDSDK_ACTIVE_CONFIG_NAME,
// CLOUDSDK_<SECTION>_<KEY>, …) and the user's home directory. Passing it as a
// value lets each invocation of an embedded gcloud-go have its own
// environment rather than sharing the process's.
type Env struct {
	// Getenv returns the value of an environment variable, or "" if unset.
	Getenv func(key string) string
	// HomeDir returns the user's home directory.
	HomeDir func() (string, error)
}

// OSEnv returns the Env of the running process.
func OSEnv() Env {
	return Env{Getenv: os.Getenv, HomeDir: os.UserHomeDir}
}

// MapEnv returns an Env whose variables come from vars and whose home
// directory is home.
func MapEnv(vars map[string]string, home string) Env {
	return Env{
		Getenv:  func(key string) string { return vars[key] },
		HomeDir: func() (string, error) { return home, nil },
	}
}
