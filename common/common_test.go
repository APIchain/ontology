package common

import (
	"testing"

	"github.com/Ontology/common/log"
)

func init() {
	log.Init(log.Path, log.Stdout)
}

func TestToCodeHash(t *testing.T) {
	tests := []struct {
		name string
		code []byte
	}{
		{"test1", []byte{1,0,2,4}},
		{"test1", []byte("awdawdawdadwawd")},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToCodeHash(tt.code)
			t.Logf("%v ToCodeHash() = %v", tt.code, got)
		})
	}
}

func TestGetNonce(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"test1"},
		{"test2"},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetNonce()
			t.Logf("GetNonce() = %v", got)
		})
	}
}

func TestIntToBytes(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{"test1", 1},
		{"test2", 258},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntToBytes(tt.n)
			t.Logf("%v IntToBytes() = %v", tt.n, got)

		})
	}
}

func TestBytesToInt16(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
	}{
		{"test1", []byte{0, 1}},
		{"test2", []byte{1, 1}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToInt16(tt.b)
			t.Logf("%v BytesToInt16() = %v", tt.b, got)
		})
	}
}

func TestIsEqualBytes(t *testing.T) {
	type args struct {
		b1 []byte
		b2 []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{[]byte{0, 1}, []byte{0, 1}}, true},
		{"test2", args{[]byte{0, 1}, []byte{1, 1}}, false},
		//{"test3",args{[]byte{1,1},[]byte{1,1}},false},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEqualBytes(tt.args.b1, tt.args.b2); got != tt.want {
				t.Errorf("IsEqualBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToHexString(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"test1", []byte{0, 1, 0, 0}},
		{"test2", []byte{1, 1, 244, 45}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToHexString(tt.data)
			t.Logf("%v ToHexString() = %v", tt.data, got)
		})
	}
}

func TestHexToBytes(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"test1", "00010000"},
		{"test2", "0101f42d"},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexToBytes(tt.value)
			if err != nil {
				t.Errorf("HexToBytes ERROR, %v", err)
			}
			t.Logf("%v HexToBytes() = %v", tt.value, got)
		})
	}
}

func TestBytesReverse(t *testing.T) {
	tests := []struct {
		name string
		u    []byte
	}{
		{"test1", []byte{0, 1, 0, 0}},
		{"test2", []byte{1, 1, 244, 45}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Before BytesReverse() = %v", tt.u)
			BytesReverse(tt.u)
			t.Logf("After BytesReverse() = %v", tt.u)
		})
	}
}

func TestHexToBytesReverse(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name  string
		value string
	}{
		{"test1", "00010000"},
		{"test2", "0101f42d"},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HexToBytesReverse(tt.value)
			if err != nil {
				t.Errorf("HexToBytesReverse ERROR, %v", err)
			}
			t.Logf("%v HexToBytesReverse() = %v", tt.value, got)
		})
	}
}

func TestClearBytes(t *testing.T) {
	type args struct {
		arr []byte
		len int
	}
	tests := []struct {
		name string
		args args
	}{
		{"test1", args{[]byte{0, 1, 0, 0}, 4}},
		{"test2", args{[]byte{1, 1, 244, 45}, 2}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Before clear is, %v", tt.args)
			ClearBytes(tt.args.arr, tt.args.len)
			t.Logf("After clear is, %v", tt.args)
		})
	}
}

func TestGetUint16Array(t *testing.T) {
	tests := []struct {
		name   string
		source []byte
	}{
		{"test1", []byte{0, 1, 0, 0}},
		{"test2", []byte{1, 1, 244, 45}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUint16Array(tt.source)
			if err != nil {
				t.Errorf("GetUint16Array ERROR, %v", err)
			}
			t.Logf("%v GetUint16Array() = %v", tt.source, got)
		})
	}
}

func TestToByteArray(t *testing.T) {
	tests := []struct {
		name   string
		source []uint16
	}{
		{"test1", []uint16{256, 0}},
		{"test2", []uint16{257, 11764}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToByteArray(tt.source)
			t.Logf("%v ToByteArray() = %v", tt.source, got)
		})
	}
}

func TestSliceRemove(t *testing.T) {
	type args struct {
		slice []uint32
		h     uint32
	}
	tests := []struct {
		name string
		args args
	}{
		{"test1", args{[]uint32{300, 14, 245, 45}, 245}},
		{"test2", args{[]uint32{1, 1, 0, 2}, 1}},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SliceRemove(tt.args.slice, tt.args.h)
			t.Logf("%v SliceRemove() = %v", tt.args, got)
		})
	}
}

func TestIsArrayEqual(t *testing.T) {
	type args struct {
		a []byte
		b []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{[]byte{0, 1, 0, 0}, []byte{0, 1, 0, 0}}, true},
		{"test2", args{[]byte{0, 1, 0, 0}, []byte{0, 1, 1, 0}}, false},
		//TODO:add another test case
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsArrayEqual(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsArrayEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
