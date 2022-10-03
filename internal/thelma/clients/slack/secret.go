package slackapi

import (
	"bytes"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

var tokenID = os.Getenv("TOKEN_ID")

func accessSecret(w io.Writer, name string) error {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secretmanger client: %v", err)
	}

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to access secret version: %v", err)
	}
	_, err = w.Write(result.GetPayload().GetData())
	if err != nil {
		return fmt.Errorf("failed to access secret payload: %v", err)
	}
	log.Print("Successfully accessed secret\n", name)
	return nil
}

func GetSlackToken() (*bytes.Buffer, error) {
	tokenUri := tokenID + "/versions/latest"
	buf := new(bytes.Buffer)
	err := accessSecret(buf, tokenUri)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
