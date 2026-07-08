package github

import "testing"

func TestParseRef(t *testing.T) {
	tests := []struct {
		input   string
		want    SkillRef
		wantErr bool
	}{
		{"user/repo", SkillRef{Owner: "user", Repo: "repo"}, false},
		{"user/repo/skill", SkillRef{Owner: "user", Repo: "repo", Path: "skill"}, false},
		{"user/repo/src/skills/commit", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills/commit"}, false},
		{"user/repo/src/skills", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills"}, false},
		{"invalid", SkillRef{}, true},
		{"/repo", SkillRef{}, true},
		{"user/", SkillRef{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseRef(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if *got != tt.want {
				t.Errorf("got %+v, want %+v", *got, tt.want)
			}
		})
	}
}
