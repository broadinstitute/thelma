// Package bootstrap handles initial installation for Thelma.
// It will:
// * Create skeleton config.yaml with `home` set to THELMA_HOME
// * Prompt the user for options:
//   - Prepend PATH with thelma’s bundled tools (this includes kubectl, helm, helmfile, vault client, and more)? [Y/n]
//   - Enable shell completion for Thelma commands? [Y/n]
//
// * Generate ~/.thelma/shell/completion.zsh
//   - This is trivial thanks to Cobra’s neat shell completion feature!
//
// * Generate ~/.thelma/shell/thelma.zsh, which
//   - Includes a “Warning: auto-generated by Thelma’s bootstrap process; do not manually edit this file!” comment
//   - Adds ~/.thelma/releases/current/bin to PATH
//   - Adds ~/.thelma/releases/current/tools/bin to PATH (if user answered “yes”)
//   - Enables shell completion (if user answered “yes”)
//   - source ~/.thelma/shell/completion.zsh && compdef _thelma thelma
//
// * Add the following line to ~/.zshrc if it does not exist:
//   - [ -f ~/.thelma/shell/thelma.zsh ] && source .thelma/shell/thelma.zsh
//   - .zshrc will backed up to ~/.zshrc.bak before writing changes
package bootstrap

import (
	_ "embed"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/autoupdate/releases"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/root"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/prompt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/shell"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"text/template"
)

const thelmaInitializationFile = "init.zsh"
const thelmaShellCompletionFile = "completion.zsh"

const addToolsToPathPrompt = "Prepend PATH with Thelma’s bundled tools" +
	" (Helm, kubectl, and more)?"

const enableShellCompletionPrompt = "Enable shell completion for Thelma commands?"

//go:embed templates/config.yaml.gotmpl
var configTemplate string

//go:embed templates/init.zsh.gotmpl
var thelmaInitTemplate string

type thelmaInitTemplateContext struct {
	AddToolsToPath        bool
	EnableShellCompletion bool
	CurrentReleaseSymlink string
	ShellCompletionFile   string
}

type Bootstrapper interface {
	// Bootstrap will bootstrap a thelma installation, including:
	// * creating skeleton config.yaml with `home` set to THELMA_HOME
	// * generating shell scripts that:
	//    * add `thelma` to path
	//    * optionally add thelma's bundled tools to PATH
	//    * optionally set up shell completion for thelma
	// * updating user's ~/.zshrc to source thelma's shell init script
	Bootstrap() error
}

// New returns a new Bootstrapper
func New(root root.Root, config config.Config, runner shell.Runner) Bootstrapper {
	return newWith(root, config, runner, prompt.New(), "")
}

// package-private constructor exposing additional options for testing
func newWith(root root.Root, config config.Config, runner shell.Runner, prompt prompt.Prompt, zshrcFile string) Bootstrapper {
	return &bootstrapper{
		root:           root,
		config:         config,
		shellRunner:    runner,
		prompt:         prompt,
		initFile:       path.Join(root.ShellDir(), thelmaInitializationFile),
		completionFile: path.Join(root.ShellDir(), thelmaShellCompletionFile),
		zshrcFile:      zshrcFile,
	}
}

type options struct {
	addToolsToPath        bool
	enableShellCompletion bool
}

type bootstrapper struct {
	root           root.Root
	config         config.Config
	shellRunner    shell.Runner
	prompt         prompt.Prompt
	initFile       string // ~/.thelma/shell/init.zsh
	completionFile string // ~/.thelma/shell/completion.zsh
	zshrcFile      string // custom .zshrc path, exposed as an option for testing
}

func (b *bootstrapper) Bootstrap() error {
	opts, err := b.promptUserForOptions()
	if err != nil {
		return err
	}

	if err = b.writeSkeletonConfigFile(); err != nil {
		return err
	}

	if opts.enableShellCompletion {
		if err = b.writeThelmaShellCompletionFile(); err != nil {
			return err
		}
	}

	if err = b.writeThelmaInitFile(opts); err != nil {
		return err
	}

	return b.addThelmaInitToZshrc()
}

func (b *bootstrapper) promptUserForOptions() (opts options, err error) {
	if err = b.prompt.Newline(); err != nil {
		return
	}
	opts.addToolsToPath, err = b.prompt.Confirm(addToolsToPathPrompt)
	if err != nil {
		return
	}
	opts.enableShellCompletion, err = b.prompt.Confirm(enableShellCompletionPrompt)
	if err != nil {
		return
	}
	if err = b.prompt.Newline(); err != nil {
		return
	}
	return
}

func (b *bootstrapper) addThelmaInitToZshrc() error {
	writer, err := newZshrcWriter(b.zshrcFile, b.initFile)
	if err != nil {
		return err
	}
	return writer.addThelmaInitialization()
}

func (b *bootstrapper) writeThelmaInitFile(opts options) error {
	ctx := thelmaInitTemplateContext{
		AddToolsToPath:        opts.addToolsToPath,
		EnableShellCompletion: opts.enableShellCompletion,
		CurrentReleaseSymlink: releases.CurrentReleaseSymlink(b.root),
		ShellCompletionFile:   b.completionFile,
	}
	log.Info().Msgf("Writing shell init script to %s...", b.initFile)
	return renderTemplateToFile(thelmaInitTemplate, ctx, b.initFile)
}

// run `thelma completion zsh` (this leverages Cobra's built-in shell completion support,
// see https://github.com/spf13/cobra/blob/main/shell_completions.md)
func (b *bootstrapper) writeThelmaShellCompletionFile() error {
	file := b.completionFile
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0755)
	if err != nil {
		return fmt.Errorf("error generating shell complation file %s: %v", file, err)
	}

	executable, err := utils.PathToRunningThelmaExecutable()
	if err != nil {
		return fmt.Errorf("error generating shell completion file %s: %v", file, err)
	}

	log.Info().Msgf("Writing shell completion script to %s...", file)
	err = b.shellRunner.Run(shell.Command{
		Prog: executable,
		Args: []string{"completion", "zsh"},
	}, func(options *shell.RunOptions) {
		options.Stdout = f
	})

	if err != nil {
		err = fmt.Errorf("error generating shell completion file %s: %v", file, err)
	}

	return utils.CloseWarn(f, err)
}

// write a ~/.thelma/config.yaml file that configures thelma's home directory and nothing else
func (b *bootstrapper) writeSkeletonConfigFile() error {
	configFile := config.DefaultConfigFilePath(b.root)
	exists, err := utils.FileExists(configFile)
	if err != nil {
		return fmt.Errorf("error generating skeleton Thelma config file: %v", err)
	}
	if exists {
		log.Warn().Msgf("%s exists; won't generate skeleton Thelma config file", configFile)
		return nil
	}

	ctx := struct {
		Home string
	}{
		Home: b.config.Home(),
	}

	log.Info().Msgf("Writing skeleton config file to %s...", configFile)
	return renderTemplateToFile(configTemplate, ctx, configFile)
}

func renderTemplateToFile(templateString string, ctx interface{}, file string) error {
	templateName := path.Base(file)
	tmpl, err := template.New(templateName).Parse(templateString)
	if err != nil {
		panic(fmt.Errorf("failed to parse embedded template %s: %v", templateName, err))
	}

	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error opening %s for rendering: %v", file, err)
	}

	err = tmpl.Execute(f, ctx)
	if err != nil {
		err = fmt.Errorf("failed to render file %s from template: %v", file, err)
	}

	return utils.CloseWarn(f, err)
}
