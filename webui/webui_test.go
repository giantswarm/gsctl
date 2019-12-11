package webui

import (
	"testing"
)

func TestBaseURL(t *testing.T) {
	type args struct {
		apiEndpoint string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "format okay",
			args:    args{apiEndpoint: "https://api.g8s.foobar.org"},
			want:    "https://happa.g8s.foobar.org",
			wantErr: false,
		},
		{
			name:    "bad format",
			args:    args{apiEndpoint: "https://api"},
			wantErr: true,
		},
		{
			name:    "control character in URL",
			args:    args{apiEndpoint: "https://api.foo.bar" + string(byte(0x7f))},
			wantErr: true,
		},
		{
			name:    "bad format",
			args:    args{apiEndpoint: "https://foo.g8s.foobar.org"},
			wantErr: true,
		},
		{
			name:    "not a valid URL",
			args:    args{apiEndpoint: "this is not a URL"},
			wantErr: true,
		},
		{
			name:    "Only schema",
			args:    args{apiEndpoint: "https://"},
			wantErr: true,
		},
		{
			name:    "Empty input",
			args:    args{apiEndpoint: ""},
			wantErr: true,
		},
		{
			name:    "Port number given",
			args:    args{apiEndpoint: "https://api.foo.bar:8080"},
			want:    "https://happa.foo.bar",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BaseURL(tt.args.apiEndpoint)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClusterDetailsURL(t *testing.T) {
	type args struct {
		apiEndpoint  string
		clusterID    string
		organization string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normal case",
			args: args{
				apiEndpoint:  "https://api.g8s.mydomain.org",
				clusterID:    "dah45",
				organization: "acme",
			},
			want: "https://happa.g8s.mydomain.org/organizations/acme/clusters/dah45",
		},
		{
			name: "error case",
			args: args{
				apiEndpoint:  "foo bar",
				clusterID:    "dah45",
				organization: "acme",
			},
			wantErr: true,
		},
		{
			name: "error case",
			args: args{
				apiEndpoint:  "https://api.g8s.foo.bar",
				clusterID:    "",
				organization: "acme",
			},
			wantErr: true,
		},
		{
			name: "error case",
			args: args{
				apiEndpoint:  "https://api.g8s.foo.bar",
				clusterID:    "mycluster",
				organization: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ClusterDetailsURL(tt.args.apiEndpoint, tt.args.clusterID, tt.args.organization)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClusterDetailsURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ClusterDetailsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
