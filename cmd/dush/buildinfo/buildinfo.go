package buildinfo

// isTestBuild is set at compile time using -ldflags.
// It should be "true" for test builds and "false" for normal builds.
var isTestBuild string = "false"

// IsTestBuild returns true if the application is built with the 'test' flag.
func IsTestBuild() bool {
	return isTestBuild == "true"
}
