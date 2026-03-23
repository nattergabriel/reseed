package github

import "testing"

func TestParseRef(t *testing.T) {
	tests := []struct {
		input   string
		want    SkillRef
		wantErr bool
	}{
		{"user/repo", SkillRef{Owner: "user", Repo: "repo"}, false},
		{"user/repo@v1.0", SkillRef{Owner: "user", Repo: "repo", Version: "v1.0"}, false},
		{"user/repo/skill", SkillRef{Owner: "user", Repo: "repo", Path: "skill"}, false},
		{"user/repo/skill@v2.0", SkillRef{Owner: "user", Repo: "repo", Path: "skill", Version: "v2.0"}, false},
		{"user/repo/src/skills/commit", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills/commit"}, false},
		{"user/repo/src/skills/commit@v1.0", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills/commit", Version: "v1.0"}, false},
		{"user/repo/src/skills", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills"}, false},
		// GitHub web URL paths - tree/<ref>/ and blob/<ref>/ are stripped
		{"user/repo/tree/main/src/skills", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills"}, false},
		{"user/repo/tree/v1.0/src/skills/commit", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills/commit"}, false},
		{"user/repo/blob/main/src/skills/commit", SkillRef{Owner: "user", Repo: "repo", Path: "src/skills/commit"}, false},
		{"user/repo/tree/main", SkillRef{Owner: "user", Repo: "repo"}, false},
		{"invalid", SkillRef{}, true},
		{"/repo", SkillRef{}, true},
		{"user/", SkillRef{}, true},
		{"user/repo@", SkillRef{}, true},
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

func TestSourceString(t *testing.T) {
	ref := SkillRef{Owner: "user", Repo: "repo"}
	got := ref.SourceString("src/skills/commit")
	want := "user/repo/src/skills/commit"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
