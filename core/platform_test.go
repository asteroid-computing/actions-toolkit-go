package core

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseLsbRelease(t *testing.T) {
	tests := []struct {
		name    string
		out     string
		want    OsVersionInfo
		wantErr bool
	}{
		{name: "ubuntu", out: "Ubuntu\n24.04\n", want: OsVersionInfo{Name: "Ubuntu", Version: "24.04"}},
		{name: "surrounding whitespace", out: "  Debian \r\n 12 \n", want: OsVersionInfo{Name: "Debian", Version: "12"}},
		{name: "release missing", out: "Ubuntu\n", wantErr: true},
		{name: "empty", out: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLsbRelease(tt.out)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseLsbRelease(%q) error = %v, wantErr %v", tt.out, err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseLsbRelease(%q) mismatch (-want +got):\n%s", tt.out, diff)
			}
		})
	}
}
