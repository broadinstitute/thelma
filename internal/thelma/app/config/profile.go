package config

import (
	"embed"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/env"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"os"
	"runtime"
)

// name for the default, empty profile used by Thelma when run interactively
const defaultProfile = "default"

// this profile is selected automatically when Thelma is _not_ run interactively
const ciProfile = "ci"

// environment variable (prefixed with THELMA_) that users can set to select a specific profile
const profileEnvVar = "CONFIG_PROFILE"

// runtime.GOOS on macs
const osx = "darwin"

//go:embed profiles/*
var profiles embed.FS

func loadProfile(options Options) ([]byte, error) {
	return loadProfileFromEmbeddedFS(chooseProfile(options))
}

func chooseProfile(options Options) string {
	if options.Profile != "" {
		return options.Profile
	}

	name := os.Getenv(env.WithEnvPrefix(profileEnvVar))
	if name != "" {
		return name
	}

	if utils.Interactive() || runtime.GOOS == osx {
		return defaultProfile
	} else {
		log.Info().Msgf("This looks like CI/CD environment; applying CI/CD configuration profile")
		return ciProfile
	}
}

func loadProfileFromEmbeddedFS(name string) ([]byte, error) {
	// default profile is empty
	if name == defaultProfile {
		return nil, nil
	}

	return profiles.ReadFile(fmt.Sprintf("profiles/%s.yaml", name))
}
