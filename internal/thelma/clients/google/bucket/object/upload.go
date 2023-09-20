package object

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"io"
	"os"
)

type Upload interface {
	SyncOperation
}

func NewUpload(fromFile string, attrs AttrSet) Upload {
	return &upload{
		file:  fromFile,
		attrs: attrs,
	}
}

type upload struct {
	file  string
	attrs AttrSet
}

func (u *upload) Kind() string {
	return "upload"
}

func (u *upload) Handler(object Object, logger zerolog.Logger) error {
	logger = addFileCtx(logger, u.file)

	fileReader, err := os.Open(u.file)
	if err != nil {
		return errors.Errorf("failed to open file: %v", err)
	}

	objWriter := object.Handle.NewWriter(object.Ctx)
	u.attrs.writeToLogEvent(logger.Debug())
	u.attrs.applyToWriter(objWriter)

	written, err := io.Copy(objWriter, fileReader)
	if err != nil {
		return errors.Errorf("write failed: %v", err)
	}
	if err = objWriter.Close(); err != nil {
		return errors.Errorf("error closing object writer: %v", err)
	}
	if err = fileReader.Close(); err != nil {
		return errors.Errorf("error closing file reader: %v", err)
	}

	logTransfer(logger, written)
	return nil
}
