package netutil

import (
	"reflect"
	"testing"
)

func TestHosts(t *testing.T) {
	type args struct {
		cidr string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "a",
			args: args{
				cidr: "192.168.1.1/32",
			},
			want: []string{
				"192.168.1.1",
			},
			wantErr: false,
		},
		{
			name: "b",
			args: args{
				cidr: "10.99.99.1/24",
			},
			want: []string{
				"192.168.1.1",
			},
			wantErr: true,
		},
		{
			name: "c",
			args: args{
				cidr: "10.99.98.1/23",
			},
			want: []string{
				"192.168.1.1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Hosts(tt.args.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hosts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Hosts() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirstIP(t *testing.T) {
	type args struct {
		cidr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "a",
			args: args{
				cidr: "192.168.4.18/24",
			},
			want:    "192.168.4.1",
			wantErr: false,
		},
		{
			name: "b",
			args: args{
				cidr: "10.99.99.123/23",
			},
			want:    "10.99.98.1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FirstIP(tt.args.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FirstIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FirstIP() got = %v, want %v", got, tt.want)
			}
		})
	}
}
