package lights

import (
	"testing"
)

func Test_rgbToColor(t *testing.T) {
	type args struct {
		r uint8
		g uint8
		b uint8
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "white",
			args: args{
				r: 255,
				g: 255,
				b: 255,
			},
			want: 16777215,
		},
		{
			name: "red",
			args: args{
				r: 255,
				g: 0,
				b: 0,
			},
			want: 16711680,
		},
		{
			name: "green",
			args: args{
				r: 0,
				g: 255,
				b: 0,
			},
			want: 65280,
		},
		{
			name: "blue",
			args: args{
				r: 0,
				g: 0,
				b: 255,
			},
			want: 255,
		},
		{
			name: "yellow",
			args: args{
				r: 0,
				g: 255,
				b: 255,
			},
			want: 65535,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rgbToColor(tt.args.r, tt.args.g, tt.args.b); got != tt.want {
				t.Errorf("rgbToColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexToColor(t *testing.T) {
	type args struct {
		hex string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{
			name: "white",
			args: args{
				hex: "FFFFFF",
			},
			want: 16777215,
		},
		{
			name: "green",
			args: args{
				hex: "00FF00",
			},
			want: 65280,
		},
		{
			name: "blue",
			args: args{
				hex: "0000FF",
			},
			want: 255,
		},
		{
			name: "red",
			args: args{
				hex: "FF0000",
			},
			want: 16711680,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HexToColor(tt.args.hex); got != tt.want {
				t.Errorf("HexToColor() = %v, want %v", got, tt.want)
			}
		})
	}
}
