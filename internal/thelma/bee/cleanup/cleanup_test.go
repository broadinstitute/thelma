package cleanup

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/providers/gitops/statefixtures"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_pubsubTopicIds(t *testing.T) {
	fixture := statefixtures.LoadFixture(statefixtures.Default, t)

	dynamic := fixture.Environment("fiab-funky-chipmunk")
	assert.ElementsMatch(t, []string{
		"leonardo-pubsub-fiab-funky-chipmunk",
		"rawls-async-import-topic-fiab-funky-chipmunk",
		"sam-group-sync-fiab-funky-chipmunk",
		"terra-fiab-funky-chipmunk-stairwaycluster-workqueue",
		"workbench-notifications-fiab-funky-chipmunk",
	}, pubsubTopicIds(dynamic))
}

func Test_projectIds(t *testing.T) {
	fixture := statefixtures.LoadFixture(statefixtures.Default, t)

	static := fixture.Environment("staging")
	assert.ElementsMatch(t, []string{"broad-dsde-staging", "terra-datarepo-staging"}, projectIds(static))

	dynamic := fixture.Environment("fiab-funky-chipmunk")
	assert.ElementsMatch(t, []string{"broad-dsde-qa"}, projectIds(dynamic))
}
