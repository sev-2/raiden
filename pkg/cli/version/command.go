package version

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
)

var versionLogger hclog.Logger = logger.HcLog()

const (
	latestVersionUrl  = "https://raiden.sev-2.com/latest.json"
	baseDownloadUrl   = "https://github.com/sev-2/raiden/releases/download"
	updateScriptShUrl = "https://raiden.sev-2.com/update.sh"
)

func Run(currentVersion string) (errUpdate error) {
	// get latest version
	latestVersion, err := FetchLatestVersion()
	if err != nil {
		return
	}

	// check if have new version
	if latestVersion == currentVersion {
		return
	}

	isUpdate, errUpdate := promptUpdateConfiguration(latestVersion)
	if errUpdate != nil {
		versionLogger.Error(errUpdate.Error())
		os.Exit(0)
	}

	if !isUpdate {
		os.Exit(0)
	}

	// ensure temporary folder
	versionLogger.Info("prepare new binary installation")
	tmpDir := fmt.Sprintf("%s/raiden", os.TempDir())
	err = ensureTmpDir(tmpDir)
	if err != nil {
		errUpdate = fmt.Errorf("error ensuring tmp directory exists: %v", err)
		return
	}

	// Determine the correct binary for the current system
	binaryName, sha256Name, err := determineBinaryName()
	if err != nil {
		errUpdate = fmt.Errorf("error determining binary name: %v", err)
		return
	}

	// download all binary
	versionLogger.Info("download new binary")
	downloadBinaryUrl := fmt.Sprintf("%s/%s/%s", baseDownloadUrl, latestVersion, binaryName)
	err = downloadFile(downloadBinaryUrl, tmpDir+"/"+binaryName)
	if err != nil {
		errUpdate = fmt.Errorf("error downloading new binary: %v", err)
		return
	}

	downloadShaBinaryUrl := fmt.Sprintf("%s/%s/%s", baseDownloadUrl, latestVersion, sha256Name)
	err = downloadFile(downloadShaBinaryUrl, tmpDir+"/"+sha256Name)
	if err != nil {
		errUpdate = fmt.Errorf("error downloading new binary sha file: %v", err)
		return
	}

	err = downloadFile(updateScriptShUrl, tmpDir+"/update.sh")
	if err != nil {
		errUpdate = fmt.Errorf("error downloading update script : %v", err)
		return
	}

	// Verify the SHA256 checksum
	versionLogger.Info("verify downloaded binary")
	valid, err := verifySHA256(tmpDir+"/"+binaryName, tmpDir+"/"+sha256Name)
	if err != nil {
		errUpdate = fmt.Errorf("error verifying SHA256 checksum: %v", err)
		return
	}
	if !valid {
		errUpdate = errors.New("sha checksum verification failed")
		return
	}

	versionLogger.Info("starting upgrade raiden in different process")
	err = updateBinary(fmt.Sprintf("%s/%s", tmpDir, "update.sh"), fmt.Sprintf("%s/%s", tmpDir, binaryName))
	if err != nil {
		versionLogger.Error(err.Error())
	}
	versionLogger.Info("follow upgrade process, bye :)")

	os.Exit(0)
	return
}

func FetchLatestVersion() (version string, errFetch error) {
	// get latest version
	resp, err := http.Get(latestVersionUrl)
	if err != nil {
		if logger.HcLog().IsInfo() {
			errFetch = errors.New("failed get latest version")
		}
		errFetch = err
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if logger.HcLog().IsInfo() {
			errFetch = errors.New("invalid version data")
		}
		errFetch = err
		return
	}

	// Parse the JSON data into the struct
	response := make(map[string]any)
	err = json.Unmarshal(body, &response)
	if err != nil {
		if logger.HcLog().IsInfo() {
			errFetch = errors.New("invalid version format data")
		}
		errFetch = err
	}

	if v, exist := response["version"]; exist {
		if vStr, isString := v.(string); isString {
			return vStr, nil
		}
	}

	return "", nil
}

func ensureTmpDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func determineBinaryName() (string, string, error) {
	var binaryName string
	var sha256Name string

	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "raiden-linux-amd64"
			sha256Name = "raiden-linux-amd64.sha256"
		case "arm64":
			binaryName = "raiden-linux-arm64"
			sha256Name = "raiden-linux-arm64.sha256"
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "raiden-macos-amd64"
			sha256Name = "raiden-macos-amd64.sha256"
		case "arm64":
			binaryName = "raiden-macos-arm64"
			sha256Name = "raiden-macos-arm64.sha256"
		}
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			binaryName = "raiden-windows-amd64-setup.exe"
			sha256Name = "raiden-windows-amd64-setup.exe.sha256"
		case "arm64":
			binaryName = "raiden-windows-arm64-setup.exe"
			sha256Name = "raiden-windows-arm64-setup.exe.sha256"
		}
	default:
		return "", "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return binaryName, sha256Name, nil
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func verifySHA256(filepath, sha256path string) (bool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))
	shaFile, err := os.ReadFile(sha256path)
	if err != nil {
		return false, err
	}

	shaContent := strings.TrimSpace(string(shaFile))
	shaParts := strings.Split(shaContent, " ")
	if len(shaParts) == 0 {
		return false, fmt.Errorf("invalid SHA256 file format")
	}
	expectedHash := shaParts[0]
	return fileHash == expectedHash, nil
}

func promptUpdateConfiguration(latestVersion string) (bool, error) {
	confirmationText := fmt.Sprintf("raiden-cli version %s has been released, do you want to update first ?", latestVersion)
	input := confirmation.New(confirmationText, confirmation.Undecided)
	input.DefaultValue = confirmation.Yes

	return input.RunPrompt()
}

func updateBinary(updateScriptPath, binaryPath string) error {
	if runtime.GOOS == "windows" {
		// TODO : implement this
	} else {
		// Check if the script file exists
		if _, err := os.Stat(updateScriptPath); os.IsNotExist(err) {
			return fmt.Errorf("failed update, missing update script")
		}

		// Set executable permissions on the script file
		if err := os.Chmod(updateScriptPath, 0755); err != nil {
			return fmt.Errorf("failed setting executable permissions : %s", err)
		}
		return spawnTerminalAndRunCommand(fmt.Sprintf("%s %s", updateScriptPath, binaryPath))
	}

	return nil
}

func spawnTerminalAndRunCommand(command string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "cmd", "/k", command)
	case "darwin":
		cmd = exec.Command("osascript", "-e", `tell application "Terminal" to do script "`+command+`"`)
	case "linux":
		_, err := getLinuxCommand(command)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}

func getLinuxCommand(command string) (*exec.Cmd, error) {
	// Define the list of terminal emulators and their commands
	terminals := []struct {
		name string
		args []string
	}{
		{"gnome-terminal", []string{"--", "bash", "-c", command + "; exec bash"}},
		{"konsole", []string{"-e", "bash", "-c", command + "; exec bash"}},
		{"xfce4-terminal", []string{"-e", "bash", "-c", command + "; exec bash"}},
		{"xterm", []string{"-hold", "-e", "bash", "-c", command}},
		{"lxterminal", []string{"-e", "bash", "-c", command}},
	}

	// Try each terminal emulator until one succeeds
	for _, term := range terminals {
		cmd := exec.Command(term.name, term.args...)
		if err := cmd.Start(); err == nil {
			return cmd, nil
		}
	}

	return nil, fmt.Errorf("no supported terminal emulator found")
}
