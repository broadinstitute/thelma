package bucket

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/utils/logid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

type operationLogger struct {
	operationKind      string
	objectName         string
	prefixedObjectName string
	bucketPrefix       string
	bucketName         string
	startTime          time.Time
}

func (i *operationLogger) operationStarted() {
	startTime := time.Now()

	i.startTime = startTime

	logger := i.logger()
	logger.Trace().Msgf("%s %s", i.operationKind, i.objectUrl())
}

func (i *operationLogger) operationFinished(err error) {
	var startTime time.Time
	startTime = i.startTime

	logger := i.logger()

	duration := time.Since(startTime)
	event := logger.Debug()
	event.Dur("duration", duration)

	if err != nil {
		event.Str("status", "error")
		event.Err(err)
		returnErr := fmt.Errorf("%s %q failed: %v", i.operationKind, i.objectUrl(), err)
		event.Msgf(returnErr.Error())
	}

	event.Str("status", "ok")
	event.Msgf("%s finished in %s", i.operationKind, duration)
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
