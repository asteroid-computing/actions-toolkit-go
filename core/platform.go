package core

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// OsVersionInfo holds operating system name and version information.
type OsVersionInfo struct {
	Name    string
	Version string
}

// GetWindowsOsInfo retrieves the Windows [OsVersionInfo] by executing PowerShell commands.
func GetWindowsOsInfo() (OsVersionInfo, error) {
	windowsVersionCmd := exec.Command("powershell", "-command", "\"(Get-CimInstance -ClassName Win32_OperatingSystem).Version\"")
	versionStr, err := windowsVersionCmd.Output()
	if err != nil {
		return OsVersionInfo{}, fmt.Errorf("error getting Windows OS Version Info: %w", err)
	}

	windowsNameCmd := exec.Command("powershell", "-command", "\"(Get-CimInstance -ClassName Win32_OperatingSystem).Caption\"")
	name, err := windowsNameCmd.Output()
	if err != nil {
		return OsVersionInfo{}, fmt.Errorf("error getting Windows OS Name Info: %w", err)
	}

	return OsVersionInfo{
		Name:    strings.TrimSpace(string(name)),
		Version: strings.TrimSpace(string(versionStr)),
	}, nil
}

// GetMacOsInfo retrieves the macOS [OsVersionInfo] by executing sw_vers commands.
func GetMacOsInfo() (OsVersionInfo, error) {
	nameInfoCmd := exec.Command("sw_vers", "-productName")
	nameStr, err := nameInfoCmd.Output()
	if err != nil {
		return OsVersionInfo{}, fmt.Errorf("error getting macOS Name Info: %w", err)
	}

	versionInfoCmd := exec.Command("sw_vers", "-productVersion")
	versionStr, err := versionInfoCmd.Output()
	if err != nil {
		return OsVersionInfo{}, fmt.Errorf("error getting macOS Version Info: %w", err)
	}

	return OsVersionInfo{
		Name:    strings.TrimSpace(string(nameStr)),
		Version: strings.TrimSpace(string(versionStr)),
	}, nil
}

// GetLinuxOsInfo retrieves the Linux [OsVersionInfo] by executing lsb_release commands.
func GetLinuxOsInfo() (OsVersionInfo, error) {
	infoCmd := exec.Command("lsb_release", "-i", "-r", "-s")
	infoStr, err := infoCmd.Output()
	if err != nil {
		return OsVersionInfo{}, fmt.Errorf("error getting Linux OS Info: %w", err)
	}

	return parseLsbRelease(string(infoStr))
}

// parseLsbRelease parses the output of `lsb_release -i -r -s`: the distributor ID and the release, one per line.
// It returns an error rather than panicking when either line is missing.
func parseLsbRelease(out string) (OsVersionInfo, error) {
	parts := strings.Split(strings.TrimSpace(out), "\n")
	if len(parts) < 2 {
		return OsVersionInfo{}, fmt.Errorf("error getting Linux OS Info: unexpected lsb_release output %q", out)
	}

	return OsVersionInfo{
		Name:    strings.TrimSpace(parts[0]),
		Version: strings.TrimSpace(parts[1]),
	}, nil
}

// PlatformInfo holds detailed information about the current platform.
type PlatformInfo struct {
	Name      string
	Platform  string
	Arch      string
	Version   string
	isWindows bool
	isMacOS   bool
	isLinux   bool
}

// GetPlatformInfo retrieves detailed [PlatformInfo] about the current platform,
// including name, platform type, architecture, and version.
func GetPlatformInfo() (PlatformInfo, error) {
	switch runtime.GOOS {
	case "windows":
		info, err := GetWindowsOsInfo()
		if err != nil {
			return PlatformInfo{}, err
		}
		return PlatformInfo{
			Name:      info.Name,
			Platform:  "windows",
			Arch:      runtime.GOARCH,
			Version:   info.Version,
			isWindows: true,
		}, nil
	case "darwin":
		info, err := GetMacOsInfo()
		if err != nil {
			return PlatformInfo{}, err
		}
		return PlatformInfo{
			Name:     info.Name,
			Platform: "macos",
			Arch:     runtime.GOARCH,
			Version:  info.Version,
			isMacOS:  true,
		}, nil
	default:
		info, err := GetLinuxOsInfo()
		if err != nil {
			return PlatformInfo{}, err
		}
		return PlatformInfo{
			Name:     info.Name,
			Platform: "linux",
			Arch:     runtime.GOARCH,
			Version:  info.Version,
			isLinux:  true,
		}, nil
	}
}
