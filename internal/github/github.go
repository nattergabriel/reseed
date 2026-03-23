package github

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Client struct {
	HTTPClient *http.Client
	Token      string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		Token:      os.Getenv("GITHUB_TOKEN"),
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	return c.HTTPClient.Do(req)
}

type tag struct {
	Name string `json:"name"`
}

// ResolveVersion resolves "latest" or empty version to the most recent tag.
func (c *Client) ResolveVersion(owner, repo, version string) (string, error) {
	if version != "" && version != VersionLatest {
		return version, nil
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags?per_page=1", owner, repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", fmt.Errorf("fetching tags for %s/%s: %w", owner, repo, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching tags for %s/%s: HTTP %d", owner, repo, resp.StatusCode)
	}

	var tags []tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return "", fmt.Errorf("parsing tags response: %w", err)
	}

	if len(tags) == 0 {
		// No tags - use default branch (empty ref in tarball URL)
		return "", nil
	}

	return tags[0].Name, nil
}

// ExtractedSkill holds the name and repo-relative path of an extracted skill.
type ExtractedSkill struct {
	Name string // directory name, e.g. "commit"
	Path string // repo-relative path, e.g. "src/skills/commit"
}

// FetchSkills downloads a repo tarball and extracts skill directories into destDir.
// If ref.Path is set, only skills at or under that path are extracted.
// Returns the extracted skills with their names and repo-relative paths.
func (c *Client) FetchSkills(ref *SkillRef, destDir string) ([]ExtractedSkill, error) {
	version, err := c.ResolveVersion(ref.Owner, ref.Repo, ref.Version)
	if err != nil {
		return nil, err
	}

	tarURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball", ref.Owner, ref.Repo)
	if version != "" {
		tarURL += "/" + version
	}

	req, err := http.NewRequest("GET", tarURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading %s/%s: %w", ref.Owner, ref.Repo, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s/%s: HTTP %d", ref.Owner, ref.Repo, resp.StatusCode)
	}

	return extractSkills(resp.Body, destDir, ref.Path)
}

// extractSkills reads a tar.gz stream and extracts skill directories.
// The tarball root is "{owner}-{repo}-{sha}/", which gets stripped.
// A skill is any directory containing a SKILL.md file.
// Skills are flattened into destDir/<skillname>/.
//
// filterPath scopes extraction:
//   - "" extracts all skills in the repo
//   - "src/skills/commit" extracts only the skill at that exact path
//   - "src/skills" extracts all skills found under that directory
func extractSkills(r io.Reader, destDir string, filterPath string) ([]ExtractedSkill, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("decompressing: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)

	type entry struct {
		name     string // path after stripping tarball root
		typeflag byte
		mode     int64
		data     []byte
	}
	var entries []entry

	// skillDirs maps "path/to/skillname/" -> skill name
	// Built by finding SKILL.md files at any depth.
	skillDirs := make(map[string]string)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}

		// Strip the root directory (e.g., "owner-repo-sha/")
		name := hdr.Name
		slashIdx := strings.Index(name, "/")
		if slashIdx == -1 {
			continue
		}
		name = name[slashIdx+1:]
		if name == "" {
			continue
		}

		var data []byte
		if hdr.Typeflag == tar.TypeReg {
			data, err = io.ReadAll(tr)
			if err != nil {
				return nil, fmt.Errorf("reading %s: %w", name, err)
			}
		}

		entries = append(entries, entry{
			name:     name,
			typeflag: hdr.Typeflag,
			mode:     hdr.Mode,
			data:     data,
		})

		// Detect SKILL.md at any depth: "some/path/skillname/SKILL.md"
		if hdr.Typeflag == tar.TypeReg && filepath.Base(name) == "SKILL.md" {
			skillDir := filepath.Dir(name)      // "some/path/skillname"
			skillName := filepath.Base(skillDir) // "skillname"

			// Apply path filter: only include skills at or under filterPath
			if filterPath != "" && skillDir != filterPath && !strings.HasPrefix(skillDir, filterPath+"/") {
				continue
			}

			skillDirs[skillDir+"/"] = skillName
		}
	}

	if filterPath != "" && len(skillDirs) == 0 {
		return nil, fmt.Errorf("no skills found under %q", filterPath)
	}

	// Extract files belonging to skill directories, flattening into destDir/skillname/
	type skillInfo struct {
		name string
		path string // repo-relative dir path
	}
	extracted := make(map[string]skillInfo)

	for _, e := range entries {
		for prefix, skillName := range skillDirs {
			if !strings.HasPrefix(e.name, prefix) && e.name != strings.TrimSuffix(prefix, "/") {
				continue
			}

			// Compute the path relative to the skill directory
			relPath := strings.TrimPrefix(e.name, prefix)
			destPath := filepath.Join(destDir, skillName, relPath)

			switch e.typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(destPath, 0o755); err != nil {
					return nil, err
				}
			case tar.TypeReg:
				if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
					return nil, err
				}
				if err := os.WriteFile(destPath, e.data, os.FileMode(e.mode).Perm()); err != nil {
					return nil, err
				}
			}

			extracted[skillName] = skillInfo{
				name: skillName,
				path: strings.TrimSuffix(prefix, "/"),
			}
			break
		}
	}

	var skills []ExtractedSkill
	for _, info := range extracted {
		skills = append(skills, ExtractedSkill{
			Name: info.name,
			Path: info.path,
		})
	}
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})
	return skills, nil
}
