package github

import (
	"fmt"
	"strings"
)

type SkillRef struct {
	Owner string
	Repo  string
	Path  string // sub-path within repo; empty = all skills
}

// String renders the ref back as user/repo[/path].
func (r SkillRef) String() string {
	s := r.Owner + "/" + r.Repo
	if r.Path != "" {
		s += "/" + r.Path
	}
	return s
}

// ParseRef parses specifiers like:
//
//	user/repo
//	user/repo/path/to/skill
func ParseRef(spec string) (SkillRef, error) {
	parts := strings.SplitN(spec, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return SkillRef{}, fmt.Errorf("invalid specifier %q: expected user/repo[/path]", spec)
	}

	ref := SkillRef{Owner: parts[0], Repo: parts[1]}
	if len(parts) == 3 {
		ref.Path = strings.TrimSuffix(parts[2], "/")
	}
	return ref, nil
}
