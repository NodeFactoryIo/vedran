package util

import "testing"

func TestIsValidPortAsInt(t *testing.T) {
	type args struct {
		port int32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns false if port negative",
			args: args{port: -1},
			want: false,
		},
		{
			name: "Returns false if port more than 49151",
			args: args{port: 49152},
			want: false,
		},
		{
			name: "Returns true if port valid",
			args: args{port: 3000},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPortAsInt(tt.args.port); got != tt.want {
				t.Errorf("IsValidPortAsInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPortAsStr(t *testing.T) {
	type args struct {
		port string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Returns false if port not a number",
			args: args{port: "invalid"},
			want: false,
		},
		{
			name: "Returns false if port negative",
			args: args{port: "-1"},
			want: false,
		},
		{
			name: "Returns false if port more than 49151",
			args: args{port: "49152"},
			want: false,
		},
		{
			name: "Returns true if port valid",
			args: args{port: "3000"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPortAsStr(tt.args.port); got != tt.want {
				t.Errorf("IsValidPortAsStr() = %v, want %v", got, tt.want)
			}
		})
	}
}
