package seed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/cli/commands/bee/seed"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

func (cmd *seedCommand) step1CreateElasticsearch(thelma app.ThelmaApp, appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("creating healthy Ontology index with Elasticsearch...")
	if elasticsearch, elasticsearchPresent := appReleases["elasticsearch"]; elasticsearchPresent {

		kubectl, err := thelma.Clients().Google().Kubectl()
		if err != nil {
			return fmt.Errorf("error getting kubectl client: %v", err)
		}

		config, err := seed.ConfigWithBasicDefaults(thelma)
		if err != nil {
			return fmt.Errorf("error getting Elasticsearch's info: %v", err)
		}

		localPort, stopFunc, err := kubectl.PortForward(elasticsearch, fmt.Sprintf("service/%s", config.Elasticsearch.Service), elasticsearch.Port())
		if err != nil {
			return fmt.Errorf("error port-forwarding to Elasticsearch: %v", err)
		}
		defer func() { _ = stopFunc() }()

		httpClient := http.Client{}
		err = _createIndex(httpClient, localPort, "ontology")
		if err = cmd.handleErrorWithForce(err); err != nil {
			return err
		}
		err = _setElasticsearchReplicas(httpClient, localPort, 0)
		if err = cmd.handleErrorWithForce(err); err != nil {
			return err
		}
	} else {
		log.Info().Msg("Elasticsearch not present in environment, skipping")
	}
	log.Info().Msg("...done")
	return nil
}

func _createIndex(client http.Client, localElasticsearchPort int, index string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("localhost:%d/%s", localElasticsearchPort, index), &bytes.Buffer{})
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error creating %s: %v", index, err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return fmt.Errorf("%s status creating %s (%s)", resp.Status, index, respBodyString)
	}
	return nil
}

func _setElasticsearchReplicas(client http.Client, localElasticsearchPort int, replicas int) error {
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
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("localhost:%d/_settings", localElasticsearchPort), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error setting replica count: %v", err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return fmt.Errorf("%s status setting replica count (%s)", resp.Status, respBodyString)
	}
	return nil
}
