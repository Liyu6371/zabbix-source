package define

import "fmt"

var (
	Version   = ""
	Commit    = ""
	BuildDate = ""
)

// PrintBuildInfo prints the build information of the application.
func PrintBuildInfo() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("BuildDate: %s\n", BuildDate)
}
