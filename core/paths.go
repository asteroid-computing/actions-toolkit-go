package core

import (
	"os"
	"strings"
)

// ToPosixPath returns pth with every backslash replaced by a forward slash, as the JS toolkit's toPosixPath does.
// Unlike filepath.ToSlash, which changes nothing outside Windows, it converts on every OS, for paths that come from a different OS than the one running the action.
// A backslash is converted even on Unix, where it is a legal filename character.
func ToPosixPath(pth string) string {
	return strings.ReplaceAll(pth, `\`, "/")
}

// ToWin32Path returns pth with every forward slash replaced by a backslash, as the JS toolkit's toWin32Path does.
// Unlike filepath.FromSlash, which changes nothing outside Windows, it converts on every OS.
func ToWin32Path(pth string) string {
	return strings.ReplaceAll(pth, "/", `\`)
}

// ToPlatformPath returns pth with every forward slash and backslash replaced by the separator of the OS running the action ([os.PathSeparator]), as the JS toolkit's toPlatformPath does.
// A backslash is converted even on Unix, where it is a legal filename character.
func ToPlatformPath(pth string) string {
	sep := string(os.PathSeparator)
	return strings.NewReplacer("/", sep, `\`, sep).Replace(pth)
}
