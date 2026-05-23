package output

import "testing"

func TestDisplayURLRedactsByDefault(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "plain host",
			raw:  "https://rpc.example.com",
			want: "https://rpc.example.com",
		},
		{
			name: "path secret",
			raw:  "https://mainnet.infura.io/v3/API_KEY",
			want: "https://mainnet.infura.io/...",
		},
		{
			name: "query string",
			raw:  "https://rpc.example.com?api_key=secret",
			want: "https://rpc.example.com",
		},
		{
			name: "fragment",
			raw:  "https://rpc.example.com#secret",
			want: "https://rpc.example.com",
		},
		{
			name: "userinfo",
			raw:  "https://user:pass@rpc.example.com",
			want: "https://rpc.example.com",
		},
		{
			name: "path query and fragment",
			raw:  "https://rpc.example.com/key/secret?token=abc#frag",
			want: "https://rpc.example.com/...",
		},
		{
			name: "root path",
			raw:  "https://rpc.example.com/",
			want: "https://rpc.example.com/",
		},
		{
			name: "invalid url",
			raw:  "://bad-url",
			want: "...",
		},
		{
			name: "missing scheme",
			raw:  "rpc.example.com/path",
			want: "...",
		},
		{
			name: "missing host",
			raw:  "https:///path",
			want: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayURL(tt.raw, false)
			if got != tt.want {
				t.Fatalf("displayURL(%q, false) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestDisplayURLShowURLsReturnsOriginal(t *testing.T) {
	raw := "https://user:pass@rpc.example.com/key/secret?token=abc#frag"

	got := displayURL(raw, true)
	if got != raw {
		t.Fatalf("displayURL(%q, true) = %q, want original URL", raw, got)
	}
}
