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
		{"user/repo/skill", SkillRef{Owner: "user", Repo: "repo", Skill: "skill"}, false},
		{"user/repo/skill@v2.0", SkillRef{Owner: "user", Repo: "repo", Skill: "skill", Version: "v2.0"}, false},
		{"invalid", SkillRef{}, true},
		{"a/b/c/d", SkillRef{}, true},
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
	got := ref.SourceString("my-skill")
	want := "user/repo/my-skill"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
