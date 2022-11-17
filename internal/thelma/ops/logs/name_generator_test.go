package logs

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_NameGenerator(t *testing.T) {
	generator := newLogNameGenerator()

	leonardo := mocks.NewRelease(t)
	leonardo.EXPECT().Name().Return("leonardo")

	sam := mocks.NewRelease(t)
	sam.EXPECT().Name().Return("sam")

	assert.Equal(t, "app.log", generator.generateName(container{
		containerName: "app",
		release:       sam,
	}))

	assert.Equal(t, "app-1.log", generator.generateName(container{
		containerName: "app",
		release:       sam,
	}))

	assert.Equal(t, "app.log", generator.generateName(container{
		containerName: "app",
		release:       leonardo,
	}))
}
