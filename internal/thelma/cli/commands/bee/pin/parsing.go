package pin

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"strings"
)

const legacyPropertySuffix = "_img"

type versionsParser func([]byte) (map[string]terra.VersionOverride, error)

var versionsParsers = map[string]versionsParser{
	"yaml":                      parseYaml,
	"json":                      parseJson,
	"legacy-jenkins-properties": parseLegacyJenkinsProperties,
}

func parseVersions(format string, input []byte) (map[string]terra.VersionOverride, error) {
	parser, exists := versionsParsers[format]
	if !exists {
		return nil, fmt.Errorf("--%s: invalid format %q (valid formats are: %s)", flagNames.versionsFormat, format, utils.QuoteJoin(versionFormats()))
	}
	return parser(input)
}

func versionFormats() []string {
	var result []string
	for name := range versionsParsers {
		result = append(result, name)
	}
	return result
}

func parseYaml(input []byte) (map[string]terra.VersionOverride, error) {
	result := make(map[string]terra.VersionOverride)
	err := yaml.Unmarshal(input, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseJson(input []byte) (map[string]terra.VersionOverride, error) {
	result := make(map[string]terra.VersionOverride)
	err := json.Unmarshal(input, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseLegacyJenkinsProperties(input []byte) (map[string]terra.VersionOverride, error) {
	nameConverters := map[string]string{
		"consent-ontology":        "ontology",
		"firecloud-orchestration": "firecloudorch",
		"firecloud-ui":            "firecloudui",
		"job-manager":             "jobmanager",
	}

	result := make(map[string]terra.VersionOverride)

	scanner := bufio.NewScanner(bytes.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) < 2 || !strings.HasSuffix(tokens[0], legacyPropertySuffix) {
			log.Warn().Msgf("Ignoring invalid line in version properties file:\n%q", line)
		}
		key := tokens[0]
		key = strings.TrimSuffix(key, legacyPropertySuffix)
		key = strings.ReplaceAll(key, "_", "-")
		match, exists := nameConverters[key]
		if exists {
			key = match
		}
		log.Debug().Msgf("Converted legacy properties key %q to service name: %s", tokens[0], key)
		value := strings.TrimSpace(tokens[1])
		log.Debug().Msgf("Image version for %s: %s", key, value)
		result[key] = terra.VersionOverride{AppVersion: value}
	}

	return result, nil
}
