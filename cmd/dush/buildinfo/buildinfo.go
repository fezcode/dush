package buildinfo

// Version of the application, set at compile time.
var Version string = "unknown"

// Commit of the application, set at compile time.
var Commit string = "unknown"

// BuildDate of the application, set at compile time.
var BuildDate string = "unknown"

// isTestBuild is set at compile time using -ldflags.
// It should be "true" for test builds and "false" for normal builds.
var isTestBuild string = "false"

// IsTestBuild returns true if the application is built with the 'test' flag.
func IsTestBuild() bool {
	return isTestBuild == "true"
}
