package gitfresh

import (
	"os"
	"os/exec"
)

type OSCommander interface {
	RunProgram(path string, workdir string, args ...string) ([]byte, error)
}
type OSPather interface {
	LookProgram(cmd string) (string, error)
}
type OSDirer interface {
	WalkDirFunc(path string, fn func(string)) error
}

type OSDirCommand interface {
	OSCommander
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
