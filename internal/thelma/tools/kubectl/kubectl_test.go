package kubectl

import "testing"

func Test_parsePortFromPortForwardOutput(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "port when normal",
			arg: `
Forwarding from 127.0.0.1:58795 -> 5432
Forwarding from [::1]:58795 -> 5432

`,
			want: 58795,
		},
		{
			name: "zero when nothing",
			arg:  "",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePortFromPortForwardOutput(tt.arg); got != tt.want {
				t.Errorf("parsePortFromPortForwardOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
