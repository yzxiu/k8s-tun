package info

import (
	"fmt"
	"net"
	"reflect"
	"testing"
)

func TestNewIpManage(t *testing.T) {
	type args struct {
		cidr string
	}
	tests := []struct {
		name string
		args args
		want IpManage
	}{
		{
			name: "a",
			args: args{
				cidr: "10.99.99.1/24",
			},
			want: &manager{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIpManage(tt.args.cidr, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewIpManage() = %v, want %v", got, tt.want)
			}
		})
	}
}



func TestNewIpManage1(t *testing.T) {

	ip1 := net.ParseIP("10.99.98.199/23")

	fmt.Println(ip1)
}
