package main

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"
)

func TestBitBucket_MakeRequestToTalkToEndpoint(t *testing.T) {
	type args struct {
		method string
		path   []string
		body   *bytes.Reader
	}
	tests := []struct {
		name  string
		setup func() *BitBucket
		args  args
		want  *http.Request
	}{
		// TODO: Add test cases.
		{
			name:  "Simple test",
			setup: func() *BitBucket { x := BitBucket{}; return &x },
			args: args{
				method: "GET",
				path:   []string{"/bob/"},
				body:   bytes.NewReader([]byte("")),
			},
			want: func() *http.Request {
				x, _ := http.NewRequest("GET", "https://x/bob/", bytes.NewReader([]byte("")))
				return x
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setup()
			if got := b.MakeRequestToTalkToEndpoint(tt.args.method, tt.args.path, tt.args.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BitBucket.MakeRequestToTalkToEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
