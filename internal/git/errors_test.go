package git

import (
	"strings"
	"testing"

	"github.com/olymp-software/chiron/internal/apperr"
)

// The outputs are captured from git 2.4x and OpenSSH with LC_ALL=C.
func TestClassify(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		want   apperr.Code
	}{
		{
			name:   "https auth failed",
			stderr: "remote: Invalid username or token. Password authentication is not supported for Git operations.\nfatal: Authentication failed for 'https://github.com/owner/private.git/'\n",
			want:   apperr.RepoAccessDenied,
		},
		{
			name:   "terminal prompts disabled",
			stderr: "fatal: could not read Username for 'https://github.com': terminal prompts disabled\n",
			want:   apperr.RepoAccessDenied,
		},
		{
			name:   "ssh publickey",
			stderr: "git@github.com: Permission denied (publickey).\r\nfatal: Could not read from remote repository.\n\nPlease make sure you have the correct access rights\nand the repository exists.\n",
			want:   apperr.RepoAccessDenied,
		},
		{
			name:   "https repository not found",
			stderr: "remote: Repository not found.\nfatal: repository 'https://github.com/owner/missing/' not found\n",
			want:   apperr.RepoNotFound,
		},
		{
			name:   "ssh repository not found",
			stderr: "ERROR: Repository not found.\nfatal: Could not read from remote repository.\n\nPlease make sure you have the correct access rights\nand the repository exists.\n",
			want:   apperr.RepoNotFound,
		},
		{
			name:   "not a repository",
			stderr: "fatal: '/srv/git/missing.git' does not appear to be a git repository\nfatal: Could not read from remote repository.\n",
			want:   apperr.RepoNotFound,
		},
		{
			name:   "http 404",
			stderr: "fatal: unable to access 'https://example.com/team/missing.git/': The requested URL returned error: 404\n",
			want:   apperr.RepoNotFound,
		},
		{
			name:   "host key verification",
			stderr: "No ED25519 host key is known for gitlab.example.com and you have requested strict checking.\r\nHost key verification failed.\r\nfatal: Could not read from remote repository.\n",
			want:   apperr.SshHostUnknown,
		},
		{
			name:   "https could not resolve host",
			stderr: "fatal: unable to access 'https://no-such-host.example/repo/': Could not resolve host: no-such-host.example\n",
			want:   apperr.RepoUnreachable,
		},
		{
			name:   "ssh could not resolve hostname",
			stderr: "ssh: Could not resolve hostname no-such-host.example: Name or service not known\r\nfatal: Could not read from remote repository.\n",
			want:   apperr.RepoUnreachable,
		},
		{
			name:   "https connection timeout",
			stderr: "fatal: unable to access 'https://10.255.255.1/repo/': Failed to connect to 10.255.255.1 port 443 after 75003 ms: Couldn't connect to server\n",
			want:   apperr.RepoUnreachable,
		},
		{
			name:   "ssh connection timeout",
			stderr: "ssh: connect to host 10.255.255.1 port 22: Connection timed out\r\nfatal: Could not read from remote repository.\n",
			want:   apperr.RepoUnreachable,
		},
		{
			name:   "network unreachable",
			stderr: "ssh: connect to host github.com port 22: Network is unreachable\r\nfatal: Could not read from remote repository.\n",
			want:   apperr.RepoUnreachable,
		},
		{
			name:   "unknown",
			stderr: "error: pathspec 'refs/tags/v9.9.9' did not match any file(s) known to git\n",
			want:   apperr.GitFailed,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classify(tt.stderr)
			if got.Code != tt.want {
				t.Errorf("classify = %s, want %s", got.Code, tt.want)
			}
			if tt.want == apperr.GitFailed && got.Params["output"] != strings.TrimSpace(tt.stderr) {
				t.Errorf("output = %q, want the error output", got.Params["output"])
			}
		})
	}
}

func TestLastLines(t *testing.T) {
	var b strings.Builder
	for i := range 15 {
		b.WriteString("line ")
		b.WriteByte(byte('a' + i))
		b.WriteString("\r\n\n")
	}
	got := lastLines(b.String(), 3)
	if want := "line m\nline n\nline o"; got != want {
		t.Errorf("lastLines = %q, want %q", got, want)
	}
}
