package server

import (
	"testing"
)

func TestAddrPool_Init(t *testing.T) {
	type fields struct {
		addrMap map[int]*RemoteID
	}

	type args struct {
		rang string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Returns error if port range format invalid",
			args:    args{rang: "i"},
			wantErr: true,
		},
		{
			name:    "Returns error if port range 1 less than range 2",
			args:    args{rang: "200:100"},
			wantErr: true,
		},
		{
			name:    "Creates addr map if port range valid",
			args:    args{rang: "100:200"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &AddrPool{
				addrMap: tt.fields.addrMap,
			}
			if err := ap.Init(tt.args.rang); (err != nil) != tt.wantErr {
				t.Errorf("AddrPool.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddrPool_Acquire(t *testing.T) {
	type fields struct {
		first   int
		last    int
		used    int
		addrMap map[int]*RemoteID
	}
	type args struct {
		cname string
		pname string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Returns error if no available ports",
			args:    args{cname: "test-id", pname: "default"},
			wantErr: true,
			want:    0,
			fields:  fields{100, 100, 1, make(map[int]*RemoteID)},
		},
		{
			name:    "Returns existing http port if exists",
			args:    args{cname: "test-id", pname: "http"},
			wantErr: false,
			want:    100,
			fields: fields{100, 102, 1, map[int]*RemoteID{100: {
				ClientID: "test-id",
				PortName: "http",
				Port:     100,
			}}},
		},
		{
			name:    "Returns existing ws port if exists",
			args:    args{cname: "test-id", pname: "ws"},
			wantErr: false,
			want:    100,
			fields: fields{100, 102, 1, map[int]*RemoteID{100: {
				ClientID: "test-id",
				PortName: "ws",
				Port:     100,
			}}},
		},
		{
			name:    "Acquires port if available",
			args:    args{cname: "test-id", pname: "default"},
			wantErr: false,
			want:    100,
			fields:  fields{100, 101, 1, make(map[int]*RemoteID)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &AddrPool{
				first:   tt.fields.first,
				last:    tt.fields.last,
				used:    tt.fields.used,
				addrMap: tt.fields.addrMap,
			}
			got, err := ap.Acquire(tt.args.cname, tt.args.pname)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddrPool.Acquire() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddrPool.Acquire() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddrPool_Release(t *testing.T) {

	addrMap := make(map[int]*RemoteID)
	addrMap[100] = &RemoteID{
		ClientID: "valid-id",
	}

	type fields struct {
		first   int
		last    int
		used    int
		addrMap map[int]*RemoteID
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Returns error if id not found in pool",
			args:    args{id: "invalid"},
			wantErr: true,
			fields:  fields{100, 100, 1, make(map[int]*RemoteID)},
		},
		{
			name:    "Returns nil if id in pool",
			args:    args{id: "valid-id"},
			wantErr: false,
			fields:  fields{100, 101, 1, addrMap},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &AddrPool{
				first:   tt.fields.first,
				last:    tt.fields.last,
				used:    tt.fields.used,
				addrMap: tt.fields.addrMap,
			}
			if err := ap.Release(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("AddrPool.Release() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddrPool_GetHTTPPort(t *testing.T) {
	addrMap := make(map[int]*RemoteID)
	addrMap[100] = &RemoteID{
		ClientID: "valid-id",
		PortName: "ws",
		Port:     20000,
	}
	addrMap[101] = &RemoteID{
		ClientID: "valid-id",
		PortName: "http",
		Port:     20001,
	}

	type fields struct {
		first   int
		last    int
		used    int
		addrMap map[int]*RemoteID
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Returns error if id not found in pool",
			args:    args{id: "invalid"},
			wantErr: true,
			want:    0,
			fields:  fields{100, 100, 1, make(map[int]*RemoteID)},
		},
		{
			name:    "Returns nil if id in pool",
			args:    args{id: "valid-id"},
			wantErr: false,
			want:    20001,
			fields:  fields{100, 101, 1, addrMap},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &AddrPool{
				first:   tt.fields.first,
				last:    tt.fields.last,
				used:    tt.fields.used,
				addrMap: tt.fields.addrMap,
			}
			got, err := ap.GetHTTPPort(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddrPool.GetHTTPPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddrPool.GetHTTPPort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddrPool_GetWSPort(t *testing.T) {
	addrMap := make(map[int]*RemoteID)
	addrMap[100] = &RemoteID{
		ClientID: "valid-id",
		PortName: "http",
		Port:     20000,
	}
	addrMap[101] = &RemoteID{
		ClientID: "valid-id",
		PortName: "ws",
		Port:     20001,
	}

	type fields struct {
		first   int
		last    int
		used    int
		addrMap map[int]*RemoteID
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "Returns error if id not found in pool",
			args:    args{id: "invalid"},
			wantErr: true,
			want:    0,
			fields:  fields{100, 100, 1, make(map[int]*RemoteID)},
		},
		{
			name:    "Returns nil if id in pool",
			args:    args{id: "valid-id"},
			wantErr: false,
			want:    20001,
			fields:  fields{100, 101, 1, addrMap},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &AddrPool{
				first:   tt.fields.first,
				last:    tt.fields.last,
				used:    tt.fields.used,
				addrMap: tt.fields.addrMap,
			}
			got, err := ap.GetWSPort(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddrPool.GetWSPort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AddrPool.GetWSPort() = %v, want %v", got, tt.want)
			}
		})
	}
}
