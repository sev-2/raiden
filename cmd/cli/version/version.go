package version

import (
	"os/exec"
	"regexp"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show application information",
		Long:  "Show application, dependencies, and device information",
		Run:   versionCmd,
	}
}

var appVersion = "1.0.0-alpha"

var appName = `
 ____           
|  _ \ __ _(_) __| | ___ _ __
| |_) / _' | |/ _' |/ _ \ '_ \ 
|  _ < (_| | | (_| |  __/ | | |
|_| \_\__,_|_|\__,_|\___|_| |_|
`

var appInformationTemplate = `
App version : %s

Dependencies Information :
- golang           : v%s
- docker           : v%s
- docker-compose   : v%s
- kubectl          : v%s

Device Information :
- Operating System : %s
- Architecture     : %s
- Cpu Number       : %v		

`

func versionCmd(cmd *cobra.Command, args []string) {

	// get dependencies information
	goVersion := getVersion("go", "version")
	dockerVersion := getVersion("docker", "--version")
	dockerComposeVersion := getVersion("docker-compose", "--version")
	kubectlVersion := getVersion("kubectl", "version", "--client")

	// get device information
	os := runtime.GOOS
	arch := runtime.GOARCH
	numCPU := runtime.NumCPU()

	print := color.New(color.FgHiYellow).PrintfFunc()
	print(appName)

	print = color.New(color.FgHiWhite).PrintfFunc()
	print(
		appInformationTemplate, appVersion,
		goVersion,
		dockerVersion,
		dockerComposeVersion,
		kubectlVersion,
		os, arch, numCPU,
	)
}

func getVersion(command string, args ...string) string {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "-"
	}

	// Use a regular expression to extract version number
	re := regexp.MustCompile(`(\d+\.\d+(\.\d+)?)`)
	versionMatches := re.FindStringSubmatch(string(output))

	if len(versionMatches) > 1 {
		return versionMatches[1]
	}

	return "-"
}
