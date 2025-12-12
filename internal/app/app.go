package app

import (
	"os"
	"path/filepath"
	"sync"
)

// App holds the application's global state.
type App struct {
	currentCWD string
}

var (
	_app  *App
	_once sync.Once
	_err  error // To store error from app initialization
)

// GetApp returns the singleton App instance.
// It ensures that the application state is initialized only once.
func GetApp() *App {
	_once.Do(func() {
		_app = &App{}
		// Initialize currentCWD with the actual OS CWD at startup
		initialCWD, err := os.Getwd()
		if err != nil {
			// fmt.Fprintf(os.Stderr, "Error getting initial working directory: %v. Defaulting to '/'.\n", err)
			initialCWD = "/" // Fallback if getting CWD fails
		}
		_app.SetCurrentDir(initialCWD) // Use the setter to initialize
	})
	return _app
}

// GetCurrentDir returns the shell's current working directory.
func (a *App) GetCurrentDir() string {
	return a.currentCWD
}

// SetCurrentDir sets the shell's current working directory.
// It performs path cleaning but does not check if the path exists or is a directory.
// This check should be done by the caller (e.g., the 'cd' builtin).
func (a *App) SetCurrentDir(path string) error {
	cleanedPath := filepath.Clean(path)
	a.currentCWD = cleanedPath
	return nil // No error for setting, validation done by caller
}
