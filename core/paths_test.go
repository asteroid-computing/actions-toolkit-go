package core

import (
	"os"
	"testing"
)

func TestPathConversion(t *testing.T) {
	sep := string(os.PathSeparator)
	tests := []struct {
		in, posix, win32, platform string
	}{
		{in: "", posix: "", win32: "", platform: ""},
		{in: "a/b/c", posix: "a/b/c", win32: `a\b\c`, platform: "a" + sep + "b" + sep + "c"},
		{in: `a\b\c`, posix: "a/b/c", win32: `a\b\c`, platform: "a" + sep + "b" + sep + "c"},
		{in: `C:\dir/mixed\file.txt`, posix: "C:/dir/mixed/file.txt", win32: `C:\dir\mixed\file.txt`, platform: "C:" + sep + "dir" + sep + "mixed" + sep + "file.txt"},
		{in: "/abs//double/", posix: "/abs//double/", win32: `\abs\\double\`, platform: sep + "abs" + sep + sep + "double" + sep},
	}
	for _, tt := range tests {
		if got := ToPosixPath(tt.in); got != tt.posix {
			t.Errorf("ToPosixPath(%q) = %q, want %q", tt.in, got, tt.posix)
		}
		if got := ToWin32Path(tt.in); got != tt.win32 {
			t.Errorf("ToWin32Path(%q) = %q, want %q", tt.in, got, tt.win32)
		}
		if got := ToPlatformPath(tt.in); got != tt.platform {
			t.Errorf("ToPlatformPath(%q) = %q, want %q", tt.in, got, tt.platform)
		}
	}
}
