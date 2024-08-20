package bucket

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/logid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

// operationLogger logs details about a given GCS operation
type operationLogger struct {
	operationKind      string
	objectName         string
	prefixedObjectName string
	bucketPrefix       string
	bucketName         string
	mutex              sync.Mutex
	startTime          time.Time
	started            bool
	finished           bool
}

func (i *operationLogger) operationStarted() {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.started {
		return
	}

	i.started = true
	i.startTime = time.Now()
	logger := i.logger()
	logger.Trace().Msgf("%s %s", i.operationKind, i.objectUrl())
}

func (i *operationLogger) operationFinished(err error) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if i.finished {
		return err
	}

	i.finished = true
	logger := i.logger()

	duration := time.Since(i.startTime)
	event := logger.Debug()
	event.Dur("duration", duration)

	if err != nil {
		event.Str("status", "error")
		event.Err(err)
		returnErr := errors.Errorf("%s %q failed: %v", i.operationKind, i.objectUrl(), err)
		event.Msg(returnErr.Error())
		return returnErr
	}

	event.Str("status", "ok")
	event.Msgf("%s finished in %s", i.operationKind, duration)
	return nil
}

func (i *operationLogger) objectUrl() string {
	return fmt.Sprintf("gs://%s/%s", i.bucketName, i.prefixedObjectName)
}

func (i *operationLogger) logger() zerolog.Logger {
	ctx := log.With().
		Interface("bucket", struct {
			Name   string `json:"name"`
			Prefix string `json:"prefix"`
		}{
			Name:   i.bucketName,
			Prefix: i.bucketPrefix,
		}).
		Interface("object", struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		}{
			Name: i.objectName,
			Url:  i.objectUrl(),
		}).
		Interface("call", struct {
			Kind string `json:"kind"`
			Id   string `json:"id"`
		}{
			Kind: i.operationKind,
			Id:   logid.NewId(),
		})

	return ctx.Logger()
}
