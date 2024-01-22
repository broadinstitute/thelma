package credentials

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsTokenProviderHappy(t *testing.T) {
	type args struct {
		tp TokenProvider
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil",
			args: args{
				tp: nil,
			},
			want: false,
		},
		{
			name: "empty",
			args: args{
				tp: &MockTokenProvider{},
			},
			want: false,
		},
		{
			name: "errors",
			args: args{
				tp: &MockTokenProvider{
					ReturnErr: true,
				},
			},
			want: false,
		},
		{
			name: "happy",
			args: args{
				tp: &MockTokenProvider{
					ReturnString: "foo",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsTokenProviderHappy(tt.args.tp), "IsTokenProviderHappy(%v)", tt.args.tp)
		})
	}
}
