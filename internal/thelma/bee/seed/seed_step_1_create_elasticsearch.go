package seed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func (s *seeder) seedStep1CreateElasticsearch(appReleases map[string]terra.AppRelease, opts SeedOptions) error {
	log.Info().Msg("creating healthy Ontology index with Elasticsearch...")
	if elasticsearch, elasticsearchPresent := appReleases["elasticsearch"]; elasticsearchPresent {
		config, err := s.configWithBasicDefaults()
		if err != nil {
			return err
		}
		localPort, stopFunc, err := s.kubectl.PortForward(elasticsearch, fmt.Sprintf("service/%s", config.Elasticsearch.Service), elasticsearch.Port())
		if err != nil {
			return errors.Errorf("error port-forwarding to Elasticsearch: %v", err)
		}
		defer func() { _ = stopFunc() }()

		httpClient := http.Client{}
		err = _createIndex(httpClient, elasticsearch.Protocol(), localPort, "ontology")
		if err = opts.handleErrorWithForce(err); err != nil {
			return err
		}
		err = _setElasticsearchReplicas(httpClient, elasticsearch.Protocol(), localPort, 0)
		if err = opts.handleErrorWithForce(err); err != nil {
			return err
		}
	} else {
		log.Info().Msg("Elasticsearch not present in environment, skipping")
	}
	log.Info().Msg("...done")
	return nil
}

func _createIndex(client http.Client, protocol string, localElasticsearchPort int, index string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s://localhost:%d/%s", protocol, localElasticsearchPort, index), &bytes.Buffer{})
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Errorf("error creating %s: %v", index, err)
	}
	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return errors.Errorf("%s status creating %s (%s)", resp.Status, index, respBodyString)
	}
	return nil
}

func _setElasticsearchReplicas(client http.Client, protocol string, localElasticsearchPort int, replicas int) error {
	bodyStruct := struct {
		Index struct {
			NumberOfReplicas int `json:"number_of_replicas"`
		} `json:"index"`
	}{}
	bodyStruct.Index.NumberOfReplicas = replicas
	body, err := json.Marshal(bodyStruct)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s://localhost:%d/_settings", protocol, localElasticsearchPort), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Errorf("error setting replica count: %v", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return errors.Errorf("%s status setting replica count (%s)", resp.Status, respBodyString)
	}
	return nil
}
