package shell

import (
	"bytes"
	"github.com/broadinstitute/thelma/internal/thelma/toolbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	tmpdir := t.TempDir()

	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "sh"
	cmd.Env = []string{"VAR1=foo"}
	cmd.Args = []string{"-c", "mkdir test-dir-$VAR1"}
	cmd.Dir = tmpdir

	if err := runner.Run(cmd); err != nil {
		t.Error(err)
	}

	// Verify that the command was run and created the directory
	testDir := path.Join(tmpdir, "test-dir-foo")
	f, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("testDir does not exist: %v", err)
	}
	if !f.IsDir() {
		t.Errorf("testDir is not directory: %v", f)
	}
}

func TestSubprocessSuccess(t *testing.T) {
	tmpdir := t.TempDir()

	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "sh"
	cmd.Env = []string{"VAR1=bar"}
	cmd.Args = []string{"-c", "mkdir test-dir-$VAR1"}
	cmd.Dir = tmpdir

	subprocess := runner.PrepareSubprocess(cmd)

	if err := subprocess.Start(); err != nil {
		t.Error(err)
	}

	if err := subprocess.Wait(); err != nil {
		t.Error(err)
	}

	// Verify that the command was run and created the directory
	testDir := path.Join(tmpdir, "test-dir-bar")
	f, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("testDir does not exist: %v", err)
	}
	if !f.IsDir() {
		t.Errorf("testDir is not directory: %v", f)
	}
}

func TestSubprocessDuplicateTermination(t *testing.T) {
	tmpdir := t.TempDir()

	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "sh"
	cmd.Env = []string{"VAR1=baz"}
	cmd.Args = []string{"-c", "mkdir test-dir-$VAR1"}
	cmd.Dir = tmpdir

	subprocess := runner.PrepareSubprocess(cmd)

	if err := subprocess.Start(); err != nil {
		t.Error(err)
	}

	if err := subprocess.Wait(); err != nil {
		t.Error(err)
	}

	if err := subprocess.Stop(); err != nil {
		t.Error(err)
	}

	if err := subprocess.Stop(); err != nil {
		t.Error(err)
	}

	// Verify that the command was run and created the directory
	testDir := path.Join(tmpdir, "test-dir-baz")
	f, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("testDir does not exist: %v", err)
	}
	if !f.IsDir() {
		t.Errorf("testDir is not directory: %v", f)
	}
}

func TestSubprocessUnawaitedTermination(t *testing.T) {
	tmpdir := t.TempDir()

	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "sh"
	cmd.Env = []string{"VAR1=boo"}
	cmd.Args = []string{"-c", "mkdir test-dir-$VAR1"}
	cmd.Dir = tmpdir

	subprocess := runner.PrepareSubprocess(cmd)

	if err := subprocess.Start(); err != nil {
		t.Error(err)
	}

	time.Sleep(1 * time.Second) // never .Wait() the process, just assume it is running

	if err := subprocess.Stop(); err != nil {
		t.Error(err)
	}

	// Verify that the command was run and created the directory
	testDir := path.Join(tmpdir, "test-dir-boo")
	f, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("testDir does not exist: %v", err)
	}
	if !f.IsDir() {
		t.Errorf("testDir is not directory: %v", f)
	}
}

func TestRunFailed(t *testing.T) {
	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "sh"
	cmd.Args = []string{"-c", "echo oops >&2 && exit 42"}
	cmd.Dir = ""

	err := runner.Run(cmd)
	if err == nil {
		t.Errorf("Expected error when running command: %v", cmd)
	}
	exitErr, ok := err.(*ExitError)
	if !assert.True(t, ok, "Expected shell.ExitError, got: %v", err) {
		t.FailNow()
	}
	assert.Equal(t, "Command \"sh -c echo oops >&2 && exit 42\" exited with status 42:\noops\n", exitErr.Error())
	assert.Equal(t, "oops\n", exitErr.Stderr)
	assert.Equal(t, 42, exitErr.ExitCode)
}

func TestRunError(t *testing.T) {
	runner := newRunner(t)
	cmd := Command{}
	cmd.Prog = "echo"
	cmd.Args = []string{"a", "b"}
	cmd.Dir = path.Join(t.TempDir(), "this-file-does-not-exist")

	err := runner.Run(cmd)
	if err == nil {
		t.Errorf("Expected error when running command: %v", cmd)
	}
	_err, ok := err.(*Error)
	if !assert.True(t, ok, "Expected shell.Error, got: %v", err) {
		t.FailNow()
	}
	assert.Regexp(t, "Command \"echo a b\" failed to start", _err.Error())
}

func TestRunWithOptions(t *testing.T) {
	runner := newRunner(t)
	var err error

	stdout := bytes.NewBuffer([]byte{})
	err = runner.Run(
		Command{
			Prog: "echo",
			Args: []string{"hello"},
		},
		func(opts *RunOptions) {
			opts.Stdout = stdout
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, "hello\n", stdout.String())

	stderr := bytes.NewBuffer([]byte{})
	err = runner.Run(
		Command{
			Prog: "ls",
			Args: []string{path.Join(t.TempDir(), "does-not-exist")},
		},
		func(opts *RunOptions) {
			opts.Stderr = stderr
		},
	)
	assert.Error(t, err)
	assert.Regexp(t, "does-not-exist.*No such file or directory", stderr.String())
}

func TestRunExpandsPathToToolbox(t *testing.T) {
	// create an alternative `ls` implementation in the toolbox and make sure it is executed instead of real `ls`
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(dir, "ls"), []byte("#!/bin/bash\necho 1234567890\n"), 0755))
	_toolbox, err := toolbox.New(dir)
	require.NoError(t, err)
	runner := NewRunner(_toolbox)

	var buf bytes.Buffer

	err = runner.Run(
		Command{
			Prog: "ls",
		},
		func(opts *RunOptions) {
			opts.Stdout = &buf
		},
	)

	require.NoError(t, err)
	assert.Equal(t, "1234567890\n", buf.String())
}

func newRunner(t *testing.T) Runner {
	dir := t.TempDir()
	_toolbox, err := toolbox.New(dir)
	require.NoError(t, err)
	return NewRunner(_toolbox)
}
