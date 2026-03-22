package github

import (
	"fmt"
	"strings"
)

type SkillRef struct {
	Owner   string
	Repo    string
	Skill   string // empty = all skills in repo
	Version string // empty = latest
}

// ParseRef parses specifiers like:
//
//	user/repo
//	user/repo@version
//	user/repo/skill
//	user/repo/skill@version
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

	parts := strings.Split(spec, "/")
	switch len(parts) {
	case 2:
		ref.Owner = parts[0]
		ref.Repo = parts[1]
	case 3:
		ref.Owner = parts[0]
		ref.Repo = parts[1]
		ref.Skill = parts[2]
	default:
		return nil, fmt.Errorf("invalid specifier %q: expected user/repo or user/repo/skill", spec)
	}

	if ref.Owner == "" || ref.Repo == "" {
		return nil, fmt.Errorf("invalid specifier: owner and repo cannot be empty")
	}

	return ref, nil
}

// SourceString returns the reseed.yaml source string, e.g. "user/repo/skill"
func (r *SkillRef) SourceString(skillName string) string {
	return fmt.Sprintf("%s/%s/%s", r.Owner, r.Repo, skillName)
}
