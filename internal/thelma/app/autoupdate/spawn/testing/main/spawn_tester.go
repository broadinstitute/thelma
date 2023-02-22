package main

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/spawn"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		log.Fatal().Msgf("Usage: %s CMD [ARGS]...", os.Args[0])
	}

	var err error

	fakeRoot := os.Getenv("FAKE_THELMA_ROOT")
	if fakeRoot == "" {
		fakeRoot, err = os.MkdirTemp(os.TempDir(), "spawn-test")
		if err != nil {
			if err != nil {
				log.Fatal().Err(err).Msgf("failed to create tmp dir")
			}
		}
	}
	log.Debug().Msgf("output will be captured in %s", fakeRoot)

	logFileName := os.Getenv("FAKE_LOGFILE_NAME")
	if logFileName == "" {
		logFileName = "spawn-tester"
	}

	thelmaRoot := root.NewAt(fakeRoot)
	if err = thelmaRoot.CreateDirectories(); err != nil {
		log.Fatal().Err(err).Msgf("failed to create fake Thelma root")
	}

	_spawn := spawn.New(root.NewAt(fakeRoot), func(options *spawn.Options) {
		options.CustomExecutable = os.Args[1]
		options.LogFileName = logFileName
	})

	err = _spawn.Spawn(os.Args[2:]...)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to spawn process")
	}
	log.Info().Msgf("spawned background process")
}
