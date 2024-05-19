package gitfresh

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
)

type OSRunner interface {
	RunProgram(path string, workdir string, args ...string) ([]byte, error)
}

type OSPather interface {
	LookProgram(cmd string) (string, error)
}

type OSDirer interface {
	WalkDirFunc(path string, fn func(string)) error
}

type OSCommander interface {
	OSRunner
	OSPather
	StartProgram(path string, args ...string) (int, error)
	UserHomePath() (string, error)
	FindProgram(pid int) (bool, error)
}

type OSDirCommand interface {
	OSRunner
	OSDirer
	OSPather
}

type AppOS struct{}

func (AppOS) RunProgram(path string, workdir string, args ...string) ([]byte, error) {
	cmd := exec.Command(path, args...)
	cmd.Dir = workdir
	return cmd.CombinedOutput()
}

func (AppOS) LookProgram(cmd string) (string, error) {
	return exec.LookPath(cmd)
}

func (AppOS) WalkDirFunc(path string, fn func(string)) error {
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, f := range dirs {
		if f.IsDir() {
			fn(f.Name())
		}
	}
	return nil
}

func (AppOS) UserHomePath() (string, error) {
	return os.UserHomeDir()
}

func (AppOS) StartProgram(path string, args ...string) (int, error) {
	cmd := exec.Command(path)
	if err := cmd.Start(); err != nil {
		slog.Info(os.Getwd())
		slog.LogAttrs(
			context.Background(),
			slog.LevelError,
			"starting program",
			slog.String("error", err.Error()),
			slog.String("path", path),
			slog.Any("args", args),
		)
		return 0, err
	}
	slog.Info("running process", "pid", cmd.Process.Pid)
	return cmd.Process.Pid, nil
}

func (AppOS) FindProgram(pid int) (bool, error) {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}
	err = process.Signal(os.Signal(syscall.Signal(0)))
	if err != nil {
		return false, err
	}
	return true, nil
}
