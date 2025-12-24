package page

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    Page
		wantErr bool
	}{
		{
			name:    "empty markdown",
			args:    args{b: []byte("")},
			want:    Page{},
			wantErr: true,
		},
		{
			name: "valid markdown",
			args: args{b: []byte(`title:my title\nkey:value\n-----------\nSome content`)},
			want: Page{Title: "my title", Attributes: map[string]string{"key": "value"}, Content: []byte(`Some content`)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
