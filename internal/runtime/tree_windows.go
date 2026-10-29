package runtime

import (
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ntResumeProcess resumes every thread of a process started suspended; Go
// does not expose the main thread handle that ResumeThread would need.
var ntResumeProcess = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtResumeProcess")

// tree is the process tree of a run: a Job Object that every external
// command is assigned to before it runs, so its descendants are in the job
// too. Closing the job handle, also when the app dies, kills the rest.
type tree struct {
	// err is a failure to create the job; the run fails with it.
	err error

	mu     sync.Mutex
	job    windows.Handle
	killed bool
	closed bool
}

func newTree() *tree {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return &tree{err: fmt.Errorf("create job object: %w", err)}
	}
	info := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info))); err != nil {
		_ = windows.CloseHandle(job)
		return &tree{err: fmt.Errorf("set job object limits: %w", err)}
	}
	return &tree{job: job}
}

// prepare makes cmd start suspended and without a console window.
func (t *tree) prepare(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW | windows.CREATE_SUSPENDED,
	}
}

// add assigns the started, still suspended cmd to the job and resumes it. A
// command that cannot be assigned, or is started after kill, is terminated.
func (t *tree) add(cmd *exec.Cmd) error {
	h, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE|windows.PROCESS_SUSPEND_RESUME,
		false, uint32(cmd.Process.Pid))
	if err != nil {
		_ = cmd.Process.Kill()
		return fmt.Errorf("open process: %w", err)
	}
	defer func() { _ = windows.CloseHandle(h) }()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		_ = windows.TerminateProcess(h, 1)
		return nil
	}
	if err := windows.AssignProcessToJobObject(t.job, h); err != nil {
		_ = windows.TerminateProcess(h, 1)
		return fmt.Errorf("assign process to job object: %w", err)
	}
	if status, _, _ := ntResumeProcess.Call(uintptr(h)); status != 0 {
		_ = windows.TerminateProcess(h, 1)
		return fmt.Errorf("resume process: NTSTATUS %#x", status)
	}
	return nil
}

// release does nothing: the job forgets processes that have exited.
func (t *tree) release(*exec.Cmd) {}

// kill terminates every process in the job.
func (t *tree) kill() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.killed {
		return
	}
	t.killed = true
	if t.err == nil {
		_ = windows.TerminateJobObject(t.job, 1)
	}
}

// close kills what is left when the script exits and releases the job.
func (t *tree) close() {
	t.kill()
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed || t.err != nil {
		return
	}
	t.closed = true
	_ = windows.CloseHandle(t.job)
}
