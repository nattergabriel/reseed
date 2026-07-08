package github

import (
	"fmt"
	"strings"
)

const VersionLatest = "latest"

type SkillRef struct {
	Owner   string
	Repo    string
	Path    string // sub-path within repo; empty = all skills
	Version string // empty = latest
}

// ParseRef parses specifiers like:
//
//	user/repo
//	user/repo@version
//	user/repo/path/to/skill
//	user/repo/path/to/skills@version
func ParseRef(spec string) (*SkillRef, error) {
	ref := &SkillRef{}

	// Split off @version first
	if idx := strings.LastIndex(spec, "@"); idx != -1 {
		ref.Version = spec[idx+1:]
		spec = spec[:idx]
		if ref.Version == "" {
			return nil, fmt.Errorf("empty version in %q", spec)
		}
	}

	parts := strings.SplitN(spec, "/", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid specifier %q: expected user/repo[/path]", spec)
	}

	ref.Owner = parts[0]
	ref.Repo = parts[1]
	if len(parts) == 3 && parts[2] != "" {
		ref.Path = strings.TrimSuffix(parts[2], "/")
	}

	if ref.Owner == "" || ref.Repo == "" {
		return nil, fmt.Errorf("invalid specifier: owner and repo cannot be empty")
	}

	return ref, nil
}
