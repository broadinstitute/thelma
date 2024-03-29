package bootstrap

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/name"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/broadinstitute/thelma/internal/thelma/utils/prompt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"os"
	"path"
	"text/template"
	"time"
)

const zshrcFile = ".zshrc"

//go:embed templates/zshrc.fragment.gotmpl
var zshrcFragmentTemplate string

type zshrcWriter interface {
	addThelmaInitialization() error
}

func newZshrcWriter(zshrcPath string, thelmaInitFile string, prompt prompt.Prompt) (zshrcWriter, error) {
	if zshrcPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Errorf("could not identify user home directory: %v", err)
		}
		zshrcPath = path.Join(homeDir, zshrcFile)
	}

	return &zshrcWriterImpl{
		file:           zshrcPath,
		thelmaInitFile: thelmaInitFile,
		prompt:         prompt,
	}, nil
}

type zshrcWriterImpl struct {
	file           string
	thelmaInitFile string
	prompt         prompt.Prompt
}

func (z *zshrcWriterImpl) addThelmaInitialization() error {
	fragment, err := z.renderZshrcFragment()
	if err != nil {
		return err
	}

	if err = z.createEmptyZshrcIfDoesNotExist(); err != nil {
		return err
	}

	alreadyUpdated, err := z.alreadyContainsThelmaInitialization(fragment)
	if err != nil || alreadyUpdated {
		return err
	}

	if err = z.backupZshrc(); err != nil {
		return err
	}

	return z.appendFragmentToZshrc(fragment)
}

// render thelma shell initialization fragment that we will add to zshrc
func (z *zshrcWriterImpl) renderZshrcFragment() ([]byte, error) {
	zshrcTemplate, err := template.New(zshrcFile).Parse(zshrcFragmentTemplate)
	if err != nil {
		panic(errors.Errorf("failed to parse template for %s: %v", zshrcFile, err))
	}

	ctx := struct {
		ShellInitializationFile string
	}{
		ShellInitializationFile: z.thelmaInitFile,
	}

	var buf bytes.Buffer
	err = zshrcTemplate.Execute(&buf, ctx)
	if err != nil {
		return nil, errors.Errorf("error rendering template for %s: %v", zshrcFile, err)
	}

	return buf.Bytes(), nil
}

// create empty ~/.zshrc if it doesn't exist
func (z *zshrcWriterImpl) createEmptyZshrcIfDoesNotExist() error {
	exists, err := utils.FileExists(z.file)
	if err != nil {
		return errors.Errorf("error adding thelma initialization to %s: %v", z.file, err)
	}

	if !exists {
		log.Warn().Msgf("%s does not exist; creating empty file", z.file)

		if err = os.WriteFile(z.file, []byte{}, 0644); err != nil {
			return errors.Errorf("error adding thelma initialization to %s: %v", z.file, err)
		}
	}

	return nil
}

// checks if ~/.zshrc already contains thelma initialization and returns false if not
func (z *zshrcWriterImpl) alreadyContainsThelmaInitialization(fragment []byte) (bool, error) {
	// read zshrc and scan to see if it already includes the fragment
	content, err := os.ReadFile(z.file)
	if err != nil {
		return false, errors.Errorf("error adding thelma initialization to %s: %v", z.file, err)
	}

	if bytes.Contains(content, fragment) {
		log.Info().Msgf("%s already includes Thelma initialization; won't update", z.file)
		return true, nil
	}

	if bytes.Contains(content, []byte(name.Name)) {
		log.Warn().Msgf("%s contains a reference to %s; won't update. "+
			"Consider adding the following lines manually:", z.file, name.Name)
		if err = z.prompt.Print(string(fragment), func(opts *prompt.PrintOptions) {
			opts.Wrap = false
		}); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// back up existing zshrc before updating
func (z *zshrcWriterImpl) backupZshrc() error {
	content, err := os.ReadFile(z.file)
	if err != nil {
		return errors.Errorf("error backing up %s: %v", z.file, err)
	}

	backupFile := fmt.Sprintf("%s.%s", z.file, time.Now().Format("20060102.150405"))
	log.Info().Msgf("Backing up %s to %s", z.file, backupFile)

	if err = os.WriteFile(backupFile, content, 0644); err != nil {
		return errors.Errorf("error backing up %s: %v", z.file, err)
	}

	return nil
}

// append thelma init fragment to zshrc
func (z *zshrcWriterImpl) appendFragmentToZshrc(fragment []byte) error {
	f, err := os.OpenFile(z.file, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return errors.Errorf("error updating %s: %v", z.file, err)
	}

	log.Info().Msgf("Adding Thelma initialization to %s", z.file)
	_, err = f.Write(fragment)
	if err != nil {
		err = errors.Errorf("error updating %s: %v", z.file, err)
	}

	return utils.CloseWarn(f, err)
}
