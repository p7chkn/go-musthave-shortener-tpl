package shortener

import (
	"net/url"
	"testing"
)

func TestGetURL(t *testing.T) {
	type args struct {
		shortURL string
		longURL  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Positive test",
			args: args{
				longURL:  "http://iloverestaurant.ru/",
				shortURL: "98fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
			want:    "http://iloverestaurant.ru/",
			wantErr: false,
		},
		{
			name: "Negative test",
			args: args{
				longURL:  "http://iloverestaurant.ru/",
				shortURL: "h12398fv58Wr3hGGIzm2-aH2zA628Ng=",
			},
			want:    "",
			wantErr: true,
		},
	}

	data := url.Values{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = AddURL(tt.args.longURL, data)
			got, err := GetURL(tt.args.shortURL, data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
