package pinflags

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"regexp"
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
		return nil, errors.Errorf("--%s: invalid format %q (valid formats are: %s)", flagNames.versionsFormat, format, utils.QuoteJoin(versionFormats()))
	}
	overrides, err := parser(input)
	if err != nil {
		return nil, err
	}
	return normalizeImageTags(overrides), nil
}

func versionFormats() []string {
	var result []string
	for name := range versionsParsers {
		result = append(result, name)
	}
	return result
}

var imageNameIllegalChars = regexp.MustCompile(`[^a-zA-Z0-9_.\-]`)
var imageNameStartsWithDotOrDash = regexp.MustCompile(`^[.\-]`)

func normalizeImageTags(input map[string]terra.VersionOverride) map[string]terra.VersionOverride {
	result := make(map[string]terra.VersionOverride)
	for key, override := range input {
		override.AppVersion = normalizeImageTag(override.AppVersion)
		result[key] = override
	}
	return result
}

// this is needed for backwards-compatibility with old/legacy Jenkins pipelines. Original code:
//
//	BRANCH_NAME=$(get_image_name $CONFIG)
//	REGEX_TO_REPLACE_ILLEGAL_CHARACTERS_WITH_DASHES="s/[^a-zA-Z0-9_.\-]/-/g"
//	REGEX_TO_REMOVE_DASHES_AND_PERIODS_FROM_BEGINNING="s/^[.\-]*//g"
//	IMAGE=$(echo $BRANCH_NAME | sed -e $REGEX_TO_REPLACE_ILLEGAL_CHARACTERS_WITH_DASHES -e $REGEX_TO_REMOVE_DASHES_AND_PERIODS_FROM_BEGINNING | cut -c 1-127)  # https://docs.docker.com/engine/reference/commandline/tag/#:~:text=A%20tag%20name%20must%20be,a%20maximum%20of%20128%20characters.
//
// https://github.com/broadinstitute/firecloud-develop/blob/a8573a38698890031444166320db1a857f8a0834/run-context/fiab/scripts/FiaB_configs.sh#L125
func normalizeImageTag(imageTag string) string {
	normalized := imageNameIllegalChars.ReplaceAllString(imageTag, "-")
	normalized = imageNameStartsWithDotOrDash.ReplaceAllString(normalized, "")
	if normalized != imageTag {
		log.Info().Msgf("Rewriting illegal image tag %q to %q", imageTag, normalized)
	}
	return normalized
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
