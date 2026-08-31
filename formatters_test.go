package main

import "testing"

func TestLocFmtNum(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"small", 42, "42"},
		{"hundreds", 999, "999"},
		{"thousands", 1234, "1,234"},
		{"ten thousands", 12345, "12,345"},
		{"millions", 1234567, "1,234,567"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := locFmtNum(tt.n)
			if got != tt.want {
				t.Errorf("locFmtNum(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestHexFmtDownloads(t *testing.T) {
	tests := []struct {
		name string
		n    int64
		want string
	}{
		{"zero", 0, "0"},
		{"small", 500, "500"},
		{"thousands", 5000, "5K"},
		{"hundred thousands", 150000, "150K"},
		{"millions", 1500000, "1.5M"},
		{"hundred millions", 523456789, "523.5M"},
		{"billions", 2000000000, "2.0B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hexFmtDownloads(tt.n)
			if got != tt.want {
				t.Errorf("hexFmtDownloads(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world this is long", 10, "hello worl..."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestTagCompressAge(t *testing.T) {
	tests := []struct {
		name string
		age  string
		want string
	}{
		{"days", "3 days ago", "3d ago"},
		{"weeks", "2 weeks ago", "2w ago"},
		{"months", "5 months ago", "5mo ago"},
		{"years", "2 years ago", "2y ago"},
		{"hours", "6 hours ago", "6h ago"},
		{"minutes", "15 minutes ago", "15m ago"},
		{"passthrough", "just now", "just now"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tagCompressAge(tt.age)
			if got != tt.want {
				t.Errorf("tagCompressAge(%q) = %q, want %q", tt.age, got, tt.want)
			}
		})
	}
}

func TestTodoExtractMarker(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantMarker string
		wantText   string
	}{
		{"todo with colon", "// TODO: fix this", "TODO", "fix this"},
		{"fixme", "# FIXME: broken", "FIXME", "broken"},
		{"hack", "/* HACK workaround */", "HACK", "workaround */"},
		{"xxx", "// XXX: danger", "XXX", "danger"},
		{"no marker", "// this is a comment", "", ""},
		{"todo no space", "//TODO fix auth", "TODO", "fix auth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMarker, gotText := todoExtractMarker(tt.line)
			if gotMarker != tt.wantMarker {
				t.Errorf("todoExtractMarker(%q) marker = %q, want %q", tt.line, gotMarker, tt.wantMarker)
			}
			if gotText != tt.wantText {
				t.Errorf("todoExtractMarker(%q) text = %q, want %q", tt.line, gotText, tt.wantText)
			}
		})
	}
}

func TestLocLang(t *testing.T) {
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "Go"},
		{".py", "Python"},
		{".rs", "Rust"},
		{".ts", "TypeScript"},
		{".tsx", "TSX"},
		{".unknown", ""},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := locLang(tt.ext)
			if got != tt.want {
				t.Errorf("locLang(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}

func TestCommitParseNumber(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want int
	}{
		{"simple", "42", 42},
		{"with suffix", "15 insertions", 15},
		{"zero", "0", 0},
		{"empty", "", 0},
		{"no digits", "abc", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := commitParseNumber(tt.s)
			if got != tt.want {
				t.Errorf("commitParseNumber(%q) = %d, want %d", tt.s, got, tt.want)
			}
		})
	}
}

func TestBlameIsHex(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"valid hex", "a1b2c3d4", true},
		{"all digits", "12345678", true},
		{"all letters", "abcdef", true},
		{"upper case", "ABCDEF", true},
		{"invalid char", "xyz123", false},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := blameIsHex(tt.s)
			if got != tt.want {
				t.Errorf("blameIsHex(%q) = %v, want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestParseBlameRange(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantStart int
		wantEnd   int
		wantHas   bool
	}{
		{"empty", "", 0, 0, false},
		{"single line", "42", 42, 42, true},
		{"range dash", "10-20", 10, 20, true},
		{"range comma", "10,20", 10, 20, true},
		{"invalid", "abc", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, e, has := parseBlameRange(tt.input)
			if s != tt.wantStart || e != tt.wantEnd || has != tt.wantHas {
				t.Errorf("parseBlameRange(%q) = (%d, %d, %v), want (%d, %d, %v)",
					tt.input, s, e, has, tt.wantStart, tt.wantEnd, tt.wantHas)
			}
		})
	}
}
