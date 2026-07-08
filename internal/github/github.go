package github

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nattergabriel/reseed/internal/reseed"
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

func apiError(status int, repo string) error {
	switch status {
	case http.StatusNotFound:
		return fmt.Errorf("repository %s not found (or is private)", repo)
	case http.StatusForbidden:
		return fmt.Errorf("GitHub API rate limit exceeded for %s - set GITHUB_TOKEN to authenticate", repo)
	case http.StatusUnauthorized:
		return fmt.Errorf("GitHub authentication failed for %s - check your GITHUB_TOKEN", repo)
	default:
		return fmt.Errorf("GitHub API error for %s: HTTP %d", repo, status)
	}
}

// FetchSkills downloads the repo's default-branch tarball and extracts skill
// directories into destDir, flattened as destDir/<name>/. If ref.Path is set,
// only skills at or under that path are extracted. Returns the skill names.
func (c *Client) FetchSkills(ctx context.Context, ref *SkillRef, destDir string) ([]string, error) {
	tarURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/tarball", ref.Owner, ref.Repo)

	req, err := http.NewRequestWithContext(ctx, "GET", tarURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading %s/%s: %w", ref.Owner, ref.Repo, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, apiError(resp.StatusCode, fmt.Sprintf("%s/%s", ref.Owner, ref.Repo))
	}

	return extractSkills(resp.Body, destDir, ref.Path)
}

// extractSkills streams a repo tarball into a temp directory, then finds and
// copies skills with the same domain code used everywhere else.
func extractSkills(r io.Reader, destDir, filterPath string) ([]string, error) {
	tmp, err := os.MkdirTemp("", "reseed-*")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if err := untar(r, tmp, filterPath); err != nil {
		return nil, err
	}

	root := filepath.Join(tmp, filepath.FromSlash(filterPath))
	var found []reseed.Skill
	if reseed.IsSkill(root) {
		found = []reseed.Skill{{Name: filepath.Base(root), Path: root}}
	} else if found, err = reseed.FindSkills(root); err != nil {
		return nil, err
	}
	if len(found) == 0 && filterPath != "" {
		return nil, fmt.Errorf("no skills found under %q", filterPath)
	}

	var names []string
	for _, s := range found {
		if err := reseed.CopySkill(s.Path, filepath.Join(destDir, s.Name)); err != nil {
			return nil, err
		}
		names = append(names, s.Name)
	}
	sort.Strings(names)
	return names, nil
}

// untar extracts a tar.gz stream into dst, stripping the tarball's root
// directory ("{owner}-{repo}-{sha}/"). If filterPath is set, only entries at
// or under it are extracted.
func untar(r io.Reader, dst, filterPath string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("decompressing: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		// Strip the root directory (e.g. "owner-repo-sha/")
		idx := strings.Index(hdr.Name, "/")
		if idx == -1 {
			continue
		}
		name := hdr.Name[idx+1:]
		if name == "" {
			continue
		}
		if filterPath != "" && name != filterPath && !strings.HasPrefix(name, filterPath+"/") {
			continue
		}

		path := filepath.Join(dst, filepath.FromSlash(name))
		if !strings.HasPrefix(path, dst+string(os.PathSeparator)) {
			continue // entry escapes dst (e.g. "..") — ignore it
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := writeFile(path, tr); err != nil {
				return err
			}
		}
	}
}

func writeFile(path string, r io.Reader) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()

	_, err = io.Copy(f, r)
	return err
}
