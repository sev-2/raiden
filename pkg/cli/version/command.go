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
	"golang.org/x/mod/modfile"
)

var VersionLogger hclog.Logger = logger.HcLog()

const (
	latestVersionUrl  = "https://raiden.sev-2.com/latest.json"
	baseDownloadUrl   = "https://github.com/sev-2/raiden/releases/download"
	updateScriptShUrl = "https://raiden.sev-2.com/update.sh"
)

func Run(currentVersion string) (isLatest bool, isUpdate bool, errUpdate error) {
	// get latest version
	latestVersion, err := FetchLatestVersion()
	if err != nil {
		return
	}

	// check if have new version
	isLatest = latestVersion == currentVersion
	if isLatest {
		return
	}

	confirmationText := fmt.Sprintf("raiden-cli version %s has been released, do you want to update first ?", latestVersion)
	isUpdate, errUpdate = promptUpdateConfirmation(confirmationText)
	if errUpdate != nil {
		return
	}

	if !isUpdate {
		return
	}

	// ensure temporary folder
	VersionLogger.Info("prepare new binary installation")
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
	VersionLogger.Info("download new binary")
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

	scriptFileName := "update.sh"
	if runtime.GOOS != "windows" {
		err = downloadFile(updateScriptShUrl, tmpDir+"/"+scriptFileName)
		if err != nil {
			errUpdate = fmt.Errorf("error downloading linux update script : %v", err)
			return
		}
	}

	// Verify the SHA256 checksum
	VersionLogger.Info("verify downloaded binary")
	valid, err := verifySHA256(tmpDir+"/"+binaryName, tmpDir+"/"+sha256Name)
	if err != nil {
		errUpdate = fmt.Errorf("error verifying SHA256 checksum: %v", err)
		return
	}
	if !valid {
		errUpdate = errors.New("sha checksum verification failed")
		return
	}

	if runtime.GOOS != "windows" {
		VersionLogger.Info("starting upgrade raiden in different process")
	}

	scriptPath := fmt.Sprintf("%s/%s", tmpDir, scriptFileName)
	binaryPath := fmt.Sprintf("%s/%s", tmpDir, binaryName)
	if runtime.GOOS == "windows" {
		tmpDir = fmt.Sprintf("%s\\raiden", os.TempDir())
		binaryPath = fmt.Sprintf("%s\\%s", tmpDir, binaryName)
		scriptPath = ""
	}

	err = updateBinary(scriptPath, binaryPath)
	if err != nil {
		errUpdate = err
		return
	}

	if runtime.GOOS != "windows" {
		VersionLogger.Info("follow upgrade process, bye :)")
	}
	return
}

func RunPackage(latestVersion string) error {
	version, err := getVersion("go.mod", "github.com/sev-2/raiden")
	if err != nil {
		return err
	}

	if version == latestVersion {
		return nil
	}

	confirmationText := fmt.Sprintf("raiden package version %s has been released, do you want to update package first ?", latestVersion)
	isUpdate, errUpdate := promptUpdateConfirmation(confirmationText)
	if errUpdate != nil {
		return errUpdate
	}

	if !isUpdate {
		return nil
	}

	VersionLogger.Info("start update version")
	cmd := exec.Command("go", "get", "-u", fmt.Sprintf("github.com/sev-2/raiden@%s", latestVersion))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("err update version : %s %s", err, output)
	}
	VersionLogger.Info(fmt.Sprintf("success update package to version %s", latestVersion))
	return nil
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

func promptUpdateConfirmation(confirmationText string) (bool, error) {
	input := confirmation.New(confirmationText, confirmation.Undecided)
	input.DefaultValue = confirmation.Yes

	return input.RunPrompt()
}

func updateBinary(updateScriptPath, binaryPath string) error {

	command := fmt.Sprintf("%s %s", updateScriptPath, binaryPath)
	if runtime.GOOS == "windows" {
		command = binaryPath
	} else {
		// Check if the script file exists
		if _, err := os.Stat(updateScriptPath); os.IsNotExist(err) {
			return fmt.Errorf("failed update, missing update script")
		}

		// Set executable permissions on the script file
		if err := os.Chmod(updateScriptPath, 0755); err != nil {
			return fmt.Errorf("failed setting executable permissions : %s", err)
		}
	}

	return spawnTerminalAndRunCommand(command)
}

func spawnTerminalAndRunCommand(command string) error {
	logger.HcLog().Info(command)
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = getWindowsCommand(command)
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

func getVersion(modulePath string, packageName string) (string, error) {
	// Open the go.mod file
	file, err := os.Open(modulePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the file content
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	// Parse the go.mod file
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return "", err
	}

	// Find the version for the specified package
	for _, r := range modFile.Require {
		if r.Mod.Path == packageName {
			return r.Mod.Version, nil
		}
	}

	return "", fmt.Errorf("package %s not found in go.mod", packageName)
}

func getWindowsCommand(command string) *exec.Cmd {
	VersionLogger.Info("running new raiden installer, follow installation process :)")
	return exec.Command("cmd", "/c", command)
}
