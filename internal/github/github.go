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
// A specific version string is returned as-is.
func (c *Client) ResolveVersion(owner, repo, version string) (string, error) {
	if version != "" && version != "latest" {
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
		// No tags — use default branch (empty ref in tarball URL)
		return "", nil
	}

	return tags[0].Name, nil
}

// FetchSkills downloads a repo tarball and extracts skill directories into destDir.
// If ref.Skill is set, only that skill is extracted. Returns the names of extracted skills.
func (c *Client) FetchSkills(ref *SkillRef, destDir string) ([]string, error) {
	version := ref.Version
	if version == "latest" {
		version = ""
	}

	// Resolve version if needed
	if version == "" {
		resolved, err := c.ResolveVersion(ref.Owner, ref.Repo, "")
		if err != nil {
			return nil, err
		}
		version = resolved
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

	return extractSkills(resp.Body, destDir, ref.Skill)
}

// extractSkills reads a tar.gz stream and extracts skill directories.
// The tarball root is "{owner}-{repo}-{sha}/", which gets stripped.
// A skill is any directory containing a SKILL.md file.
// Skills can be at any depth in the repo — they are flattened into destDir/<skillname>/.
func extractSkills(r io.Reader, destDir string, onlySkill string) ([]string, error) {
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

	// skillPrefixes maps "path/to/skillname/" -> "skillname"
	// Built by finding SKILL.md files at any depth.
	skillPrefixes := make(map[string]string)

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
			skillDir := filepath.Dir(name)            // "some/path/skillname"
			skillName := filepath.Base(skillDir)      // "skillname"
			prefix := skillDir + "/"                  // "some/path/skillname/"
			skillPrefixes[prefix] = skillName
		}
	}

	// Extract files belonging to skill directories, flattening into destDir/skillname/
	extracted := make(map[string]bool)
	for _, e := range entries {
		for prefix, skillName := range skillPrefixes {
			if !strings.HasPrefix(e.name, prefix) && e.name != strings.TrimSuffix(prefix, "/") {
				continue
			}

			if onlySkill != "" && skillName != onlySkill {
				continue
			}

			// Compute the path relative to the skill directory
			relPath := strings.TrimPrefix(e.name, prefix)
			if relPath == "" && e.typeflag == tar.TypeDir {
				relPath = ""
			}
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

			extracted[skillName] = true
			break
		}
	}

	if onlySkill != "" && !extracted[onlySkill] {
		return nil, fmt.Errorf("skill %q not found in repository", onlySkill)
	}

	var names []string
	for name := range extracted {
		names = append(names, name)
	}
	return names, nil
}
