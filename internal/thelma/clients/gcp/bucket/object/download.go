package object

import (
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"os"
)

type Download interface {
	Operation
}

func NewDownload(toFile string) Download {
	return &download{
		file: toFile,
	}
}

type download struct {
	file string
}

func (d *download) Kind() string {
	return "download"
}

func (d *download) Handler(object Object, logger zerolog.Logger) error {
	logger = addFileCtx(logger, d.file)

	logger.Debug().Msg("opening file for writing")
	fileWriter, err := os.Create(d.file)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	objReader, err := object.Handle.NewReader(object.Ctx)
	if err != nil {
		return fmt.Errorf("error reading object: %v", err)
	}

	written, err := io.Copy(fileWriter, objReader)
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}
	if err = objReader.Close(); err != nil {
		return fmt.Errorf("error closing object reader: %v", err)
	}
	if err = fileWriter.Close(); err != nil {
		return fmt.Errorf("error closing file writer: %v", err)
	}

	logTransfer(logger, written)
	return nil
}
