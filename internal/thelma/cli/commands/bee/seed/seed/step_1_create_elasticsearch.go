package seed

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

func (cmd *seedCommand) step1CreateElasticsearch(appReleases map[string]terra.AppRelease) error {
	log.Info().Msg("creating healthy Ontology index with Elasticsearch...")
	if elasticsearch, elasticsearchPresent := appReleases["elasticsearch"]; elasticsearchPresent {
		httpClient := http.Client{}
		err := _createIndex(httpClient, elasticsearch, "ontology")
		if err = cmd.handleErrorWithForce(err); err != nil {
			return err
		}
		err = _setElasticsearchReplicas(httpClient, elasticsearch, 0)
		if err = cmd.handleErrorWithForce(err); err != nil {
			return err
		}
	} else {
		log.Info().Msg("Elasticsearch not present in environment, skipping")
	}
	log.Info().Msg("...done")
	return nil
}

func _createIndex(client http.Client, elasticsearch terra.AppRelease, index string) error {
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s:%d/%s", elasticsearch.URL(), elasticsearch.Port(), index), &bytes.Buffer{})
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error creating %s:%d/%s: %v", elasticsearch.URL(), elasticsearch.Port(), index, err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return fmt.Errorf("%s status creating %s:%d/%s (%s)", resp.Status, elasticsearch.URL(), elasticsearch.Port(), index, respBodyString)
	}
	return nil
}

func _setElasticsearchReplicas(client http.Client, elasticsearch terra.AppRelease, replicas int) error {
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
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s:%d/_settings", elasticsearch.URL(), elasticsearch.Port()), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error setting replica count on %s:%d: %v", elasticsearch.URL(), elasticsearch.Port(), err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return err
	}
	respBodyString := string(respBody)
	if resp.StatusCode > 299 {
		return fmt.Errorf("%s status setting replica count %s:%d (%s)", resp.Status, elasticsearch.URL(), elasticsearch.Port(), respBodyString)
	}
	return nil
}
