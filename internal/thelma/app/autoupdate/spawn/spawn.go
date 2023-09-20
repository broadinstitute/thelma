package spawn

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

const sentinelEnvVar = "PROCESS_IS_THELMA_SPAWN"
const stderrLogExt = ".out"
const stdoutLogExt = ".err"

type Spawn interface {
	// CurrentProcessIsSpawn returns true if the current process is a spawn process
	CurrentProcessIsSpawn() bool
	// Spawn spawns a Thelma subcommand in the background, detached
	// so that it lives even after the currently running process exits.
	// Its primary use case is Thelma's self-update feature.
	Spawn(args ...string) error
}

type Option func(*Options)

type Options struct {
	// LogFileName if set to non-empty value, stdout and stderr for the sub-process
	// will be written to ~/.thelma/logs/<NAME>.out and ~/.thelma/logs/<NAME>.err
	LogFileName string
	// CustomExecutable FOR USE IN TESTS ONLY use a custom executable instead of "thelma"
	CustomExecutable string
}

func New(root root.Root, opts ...Option) Spawn {
	return &spawn{
		logsDir: root.LogDir(),
		options: asOptions(opts...),
	}
}

type spawn struct {
	logsDir string
	options Options
}

func (s *spawn) CurrentProcessIsSpawn() bool {
	return os.Getenv(sentinelEnvVar) != ""
}

func (s *spawn) Spawn(args ...string) error {
	if s.CurrentProcessIsSpawn() {
		// panic since this is a bug - spawned processes should not try to launch their own spawn
		panic(errors.Errorf("won't spawn child process %q, current process is already a Thelma spawn", strings.Join(args, " ")))
	}

	// launch a `thelma update` command in the background
	// ref: https://groups.google.com/g/golang-nuts/c/shST-SDqIp4
	executable, err := s.getExecutable()
	if err != nil {
		return err
	}

	desc := cmdDescription(executable, args)
	log.Debug().Msgf("preparing to launch new background process: %q", desc)
	cmd := exec.Command(executable, args...)

	// close stdin since we don't want the child to inherit
	cmd.Stdin = nil

	if err = s.configureLogging(cmd); err != nil {
		return errors.Errorf("errof configuring logging for background process %q: %v", desc, err)
	}

	// add our sentinel env var to the environment
	env := os.Environ()
	env = append(env, fmt.Sprintf("%s=%s", sentinelEnvVar, "true"))
	cmd.Env = env

	// make sure sub-process is not part of this process group, so that if this
	// process dies or is ctrl-C'd, the background process continues executing
	// https://stackoverflow.com/questions/35433741/in-golang-prevent-child-processes-to-receive-signals-from-calling-process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pgid:    0,
		Setpgid: true,
	}

	if err = cmd.Start(); err != nil {
		return errors.Errorf("error starting background process %q: %v", desc, err)
	}
	pid := cmd.Process.Pid

	//  https://stackoverflow.com/questions/23031752/start-a-process-in-go-and-detach-from-it
	if err = cmd.Process.Release(); err != nil {
		return errors.Errorf("error detaching background process %q: %v", desc, err)
	}

	log.Debug().Msgf("%q started (pid %d) in background", desc, pid)
	return nil
}

func (s *spawn) configureLogging(cmd *exec.Cmd) error {
	if s.options.LogFileName == "" {
		// if logging not enabled, send stdout/stderr to dev/null
		cmd.Stdout = nil
		cmd.Stderr = nil
		return nil
	}

	// logging enabled, configure the command to write to ~/.thelma/logs/<NAME>.out (and .err)
	if err := s.openLogFileAndSaveTo(stdoutLogExt, &cmd.Stdout); err != nil {
		return err
	}
	if err := s.openLogFileAndSaveTo(stderrLogExt, &cmd.Stderr); err != nil {
		return err
	}
	return nil
}

func (s *spawn) openLogFileAndSaveTo(ext string, setme *io.Writer) error {
	name := s.options.LogFileName + ext
	file := path.Join(s.logsDir, name)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return errors.Errorf("error opening %s for writing: %v", file, err)
	}
	*setme = f
	return nil
}

func (s *spawn) getExecutable() (string, error) {
	if s.options.CustomExecutable != "" {
		return s.options.CustomExecutable, nil
	}
	return utils.PathToRunningThelmaExecutable()
}

func cmdDescription(executable string, args []string) string {
	var s []string
	s = append(s, executable)
	s = append(s, args...)
	return strings.Join(s, " ")
}

func asOptions(opts ...Option) Options {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
