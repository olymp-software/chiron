package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mvdan.cc/sh/v3/interp"
)

// utility is one of the app's guaranteed commands. It reports failures
// through c and keeps going where the GNU counterpart would.
type utility func(c *call, args []string)

var utilities = map[string]utility{
	"cat":   cat,
	"cp":    cp,
	"mkdir": mkdir,
	"mv":    mv,
	"rm":    rm,
	"touch": touch,
}

// utilitiesHandler runs the guaranteed commands in Go on every OS, so they
// behave the same everywhere; any other command goes to next.
func utilitiesHandler(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
	return func(ctx context.Context, args []string) error {
		u, ok := utilities[args[0]]
		if !ok {
			return next(ctx, args)
		}
		hc := interp.HandlerCtx(ctx)
		c := &call{name: args[0], dir: hc.Dir, stdin: hc.Stdin, stdout: hc.Stdout, stderr: hc.Stderr}
		u(c, args[1:])
		if c.failed {
			return interp.ExitStatus(1)
		}
		return nil
	}
}

// call is the state of one utility invocation.
type call struct {
	name   string
	dir    string
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	failed bool
}

func (c *call) fail(format string, args ...any) {
	_, _ = fmt.Fprintf(c.stderr, c.name+": "+format+"\n", args...)
	c.failed = true
}

// failPath reports err for the operand name as typed in the script, with an
// OS-independent reason for common errors.
func (c *call) failPath(name string, err error) {
	c.fail("%s: %s", name, reason(err))
}

// path resolves an operand against the script's current directory.
func (c *call) path(name string) string {
	if filepath.IsAbs(name) {
		return name
	}
	return filepath.Join(c.dir, name)
}

// flags splits args into flags and operands. Flags may appear anywhere
// before "--" and may be combined ("-rf"); only letters in allowed are
// accepted. ok is false after an unsupported flag has been reported.
func (c *call) flags(args []string, allowed string) (set map[rune]bool, operands []string, ok bool) {
	set = map[rune]bool{}
	for i, arg := range args {
		if arg == "--" {
			return set, append(operands, args[i+1:]...), true
		}
		if len(arg) < 2 || arg[0] != '-' {
			operands = append(operands, arg)
			continue
		}
		if arg[1] == '-' {
			c.fail("unsupported option '%s'", arg)
			return nil, nil, false
		}
		for _, f := range arg[1:] {
			if !strings.ContainsRune(allowed, f) {
				c.fail("unsupported option '-%c'", f)
				return nil, nil, false
			}
			set[f] = true
		}
	}
	return set, operands, true
}

func reason(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "no such file or directory"
	case errors.Is(err, fs.ErrExist):
		return "file exists"
	case errors.Is(err, fs.ErrPermission):
		return "permission denied"
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return pathErr.Err.Error()
	}
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) {
		return linkErr.Err.Error()
	}
	return err.Error()
}

// cat [file]...: no files or "-" read stdin.
func cat(c *call, args []string) {
	_, operands, ok := c.flags(args, "")
	if !ok {
		return
	}
	if len(operands) == 0 {
		operands = []string{"-"}
	}
	for _, name := range operands {
		if name == "-" {
			if _, err := io.Copy(c.stdout, c.stdin); err != nil {
				c.failPath(name, err)
			}
			continue
		}
		if err := catFile(c.stdout, c.path(name)); err != nil {
			c.failPath(name, err)
		}
	}
}

func catFile(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if info, err := f.Stat(); err != nil {
		return err
	} else if info.IsDir() {
		return errors.New("is a directory")
	}
	_, err = io.Copy(w, f)
	return err
}

// mkdir [-p] dir...
func mkdir(c *call, args []string) {
	set, operands, ok := c.flags(args, "p")
	if !ok {
		return
	}
	if len(operands) == 0 {
		c.fail("missing operand")
		return
	}
	for _, name := range operands {
		var err error
		if set['p'] {
			err = os.MkdirAll(c.path(name), 0o777)
		} else {
			err = os.Mkdir(c.path(name), 0o777)
		}
		if err != nil {
			c.failPath(name, err)
		}
	}
}

// touch file...: creates missing files, updates the times of existing ones.
func touch(c *call, args []string) {
	_, operands, ok := c.flags(args, "")
	if !ok {
		return
	}
	if len(operands) == 0 {
		c.fail("missing file operand")
		return
	}
	now := time.Now()
	for _, name := range operands {
		path := c.path(name)
		err := os.Chtimes(path, now, now)
		if errors.Is(err, fs.ErrNotExist) {
			var f *os.File
			if f, err = os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o666); err == nil {
				err = f.Close()
			}
		}
		if err != nil {
			c.failPath(name, err)
		}
	}
}

// rm [-r] [-f] path...: -f ignores missing paths and clears the read-only
// attribute before removing.
func rm(c *call, args []string) {
	set, operands, ok := c.flags(args, "rf")
	if !ok {
		return
	}
	if len(operands) == 0 && !set['f'] {
		c.fail("missing operand")
		return
	}
	for _, name := range operands {
		if base := filepath.Base(name); base == "." || base == ".." {
			c.fail("refusing to remove '.' or '..' directory: skipping '%s'", name)
			continue
		}
		path := c.path(name)
		info, err := os.Lstat(path)
		if err != nil {
			if !set['f'] || !errors.Is(err, fs.ErrNotExist) {
				c.failPath(name, err)
			}
			continue
		}
		if info.IsDir() && !set['r'] {
			c.fail("%s: is a directory", name)
			continue
		}
		if set['f'] {
			makeWritable(path)
		}
		if info.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}
		if err != nil {
			c.failPath(name, err)
		}
	}
}

// makeWritable adds owner write permission to path and everything under it.
// Errors are ignored: the removal that follows reports what still fails.
func makeWritable(path string) {
	_ = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		mode := info.Mode().Perm() | 0o200
		if d.IsDir() {
			mode |= 0o700
		}
		if mode != info.Mode().Perm() {
			_ = os.Chmod(p, mode)
		}
		return nil
	})
}

// cp [-r] src... dst: copies into dst when it is an existing directory.
func cp(c *call, args []string) {
	set, operands, ok := c.flags(args, "r")
	if !ok {
		return
	}
	targets, ok := c.targets(operands)
	if !ok {
		return
	}
	for _, t := range targets {
		info, err := os.Stat(c.path(t.src))
		if err != nil {
			c.failPath(t.src, err)
			continue
		}
		if info.IsDir() && !set['r'] {
			c.fail("-r not specified; omitting directory '%s'", t.src)
			continue
		}
		if err := c.copy(t); err != nil {
			c.failPath(t.src, err)
		}
	}
}

// mv src... dst: moves into dst when it is an existing directory; falls back
// to copy-and-remove when the rename crosses volumes.
func mv(c *call, args []string) {
	_, operands, ok := c.flags(args, "")
	if !ok {
		return
	}
	targets, ok := c.targets(operands)
	if !ok {
		return
	}
	for _, t := range targets {
		src, dst := c.path(t.src), c.path(t.dst)
		if _, err := os.Lstat(src); err != nil {
			c.failPath(t.src, err)
			continue
		}
		err := os.Rename(src, dst)
		if err != nil && crossDevice(err) {
			if err = c.copy(t); err == nil {
				makeWritable(src)
				err = os.RemoveAll(src)
			}
		}
		if err != nil {
			c.failPath(t.src, err)
		}
	}
}

// target is one source operand and the path it is copied or moved to.
type target struct{ src, dst string }

// targets pairs the source operands with their destinations: inside the
// last operand when it is an existing directory, the last operand itself
// for a single source.
func (c *call) targets(operands []string) ([]target, bool) {
	if len(operands) < 2 {
		c.fail("missing destination file operand")
		return nil, false
	}
	srcs, dst := operands[:len(operands)-1], operands[len(operands)-1]
	info, err := os.Stat(c.path(dst))
	if err == nil && info.IsDir() {
		targets := make([]target, len(srcs))
		for i, src := range srcs {
			targets[i] = target{src, filepath.Join(dst, filepath.Base(src))}
		}
		return targets, true
	}
	if len(srcs) > 1 {
		c.fail("target '%s' is not a directory", dst)
		return nil, false
	}
	return []target{{srcs[0], dst}}, true
}

// copy copies a file or a directory tree from t.src to t.dst.
func (c *call) copy(t target) error {
	src, dst := c.path(t.src), c.path(t.dst)
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if dstInfo, err := os.Stat(dst); err == nil && os.SameFile(srcInfo, dstInfo) {
		return fmt.Errorf("'%s' and '%s' are the same file", t.src, t.dst)
	}
	if !srcInfo.IsDir() {
		return copyFile(src, dst, srcInfo.Mode().Perm())
	}
	if rel, err := filepath.Rel(src, dst); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("cannot copy a directory into itself, '%s'", t.dst)
	}
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		to := filepath.Join(dst, rel)
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return os.MkdirAll(to, info.Mode().Perm()|0o700)
		}
		return copyFile(p, to, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}
