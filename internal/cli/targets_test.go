package cli

import (
	"reflect"
	"testing"
)

func TestNormalizeTargets(t *testing.T) {
	tests := []struct {
		name      string
		requested []string
		all       bool
		defaults  []string
		want      []string
		wantErr   bool
	}{
		{
			name:     "defaults",
			defaults: []string{"codex"},
			want:     []string{"codex"},
		},
		{
			name:      "comma separated",
			requested: []string{"claude,codex"},
			want:      []string{"claude", "codex"},
		},
		{
			name:      "dedupe and case normalize",
			requested: []string{"Codex", "codex", "gemini"},
			want:      []string{"codex", "gemini"},
		},
		{
			name: "all flag",
			all:  true,
			want: []string{"claude", "gemini", "codex"},
		},
		{
			name:      "all target",
			requested: []string{"all"},
			want:      []string{"claude", "gemini", "codex"},
		},
		{
			name:      "unknown",
			requested: []string{"cursor"},
			wantErr:   true,
		},
		{
			name:    "empty without defaults",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeTargets(tt.requested, tt.all, tt.defaults)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
