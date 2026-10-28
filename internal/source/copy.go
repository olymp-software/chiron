package source

import (
	"bufio"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const (
	gitDir        = ".git"
	gitignoreFile = ".gitignore"
)

// CopyLocal copies the local course folder src into dst, which must not
// exist. It skips .git and everything excluded by .gitignore files inside src,
// matched as git matches them; .git/info/exclude and global ignore files do
// not apply. Symbolic links are not copied; their paths relative to src are
// returned so the caller can log them.
func CopyLocal(src, dst string) (skippedLinks []string, err error) {
	c := &copier{}
	if err := c.dir(src, dst, nil, nil); err != nil {
		return nil, err
	}
	return c.skippedLinks, nil
}

type copier struct {
	skippedLinks []string
}

// dir copies folder src (at parts relative to the course root) to dst. ps
// holds the ignore patterns of the parent folders.
func (c *copier) dir(src, dst string, parts []string, ps []gitignore.Pattern) error {
	own, err := readGitignore(filepath.Join(src, gitignoreFile), parts)
	if err != nil {
		return err
	}
	ps = append(ps[:len(ps):len(ps)], own...)
	matcher := gitignore.NewMatcher(ps)

	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.Mkdir(dst, info.Mode().Perm()|0o700); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == gitDir {
			continue
		}
		child := append(parts[:len(parts):len(parts)], name)
		srcPath, dstPath := filepath.Join(src, name), filepath.Join(dst, name)
		switch {
		case e.Type()&fs.ModeSymlink != 0:
			c.skippedLinks = append(c.skippedLinks, strings.Join(child, "/"))
		case e.IsDir():
			if matcher.Match(child, true) {
				continue
			}
			if err := c.dir(srcPath, dstPath, child, ps); err != nil {
				return err
			}
		case e.Type().IsRegular():
			if matcher.Match(child, false) {
				continue
			}
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// readGitignore parses the .gitignore file at path, whose folder is at parts
// relative to the course root. A missing file gives no patterns.
func readGitignore(path string, parts []string) ([]gitignore.Pattern, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var ps []gitignore.Pattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		ps = append(ps, gitignore.ParsePattern(line, parts))
	}
	return ps, scanner.Err()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
