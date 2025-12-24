package page

import (
	"reflect"
	"testing"
)

func TestSubstitute(t *testing.T) {
	type args struct {
		html []byte
		p    Page
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty template",
			args:    args{html: []byte("")},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid template",
			args:    args{html: []byte(`{{title}}{{content}}`), p: Page{Title: "title", Content: []byte(`content`)}},
			want:    []byte(`titlecontent`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Substitute(tt.args.html, tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Substitute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Substitute() got = %v, want %v", got, tt.want)
			}
		})
	}
}
