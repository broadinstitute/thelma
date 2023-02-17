package spawn

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const sentinelEnvVar = "PROCESS_IS_THELMA_SPAWN"

type Spawn interface {
	// CurrentProcessIsSpawn returns true if the current process is a spawn process
	CurrentProcessIsSpawn() bool
	// Spawn spawns a Thelma subcommand in the background, detached
	// so that it lives even after the currently running process exits.
	// Its primary use case is Thelma's self-update feature.
	Spawn(args ...string) error
}

func New() Spawn {
	return &spawn{}
}

type spawn struct {
}

func (s *spawn) CurrentProcessIsSpawn() bool {
	return os.Getenv(sentinelEnvVar) != ""
}

func (s *spawn) Spawn(args ...string) error {
	if s.CurrentProcessIsSpawn() {
		// panic since this is a bug - spawned processes should not try to launch their own spawn
		panic(fmt.Errorf("won't spawn child process %q, current process is already a Thelma spawn", strings.Join(args, " ")))
	}

	// launch a `thelma update` command in the background
	// ref: https://groups.google.com/g/golang-nuts/c/shST-SDqIp4
	executable, err := root.PathToRunningThelmaExecutable()
	if err != nil {
		return err
	}

	desc := cmdDescription(executable, args)
	log.Debug().Msgf("preparing to launch new background process: %q", desc)
	cmd := exec.Command(executable, args...)

	// stderr, stdout, stdin should all be nil/closed. (Logs will still be written to ~/.thelma/logs/)
	cmd.Stderr = nil
	cmd.Stdout = nil
	cmd.Stdin = nil

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
		return fmt.Errorf("error starting background process %q: %v", desc, err)
	}

	//  https://stackoverflow.com/questions/23031752/start-a-process-in-go-and-detach-from-it
	if err = cmd.Process.Release(); err != nil {
		return fmt.Errorf("error detaching background process %q: %v", desc, err)
	}

	log.Debug().Msgf("%q started (pid %d) in background", desc, cmd.Process.Pid)
	return nil
}

func cmdDescription(executable string, args []string) string {
	var s []string
	s = append(s, executable)
	s = append(s, args...)
	return strings.Join(s, " ")
}
