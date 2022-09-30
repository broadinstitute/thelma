package seed

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ignore409Conflict(t *testing.T) {
	err := errors.New(`
409 Conflict from https://firecloudorch.bee.envs-terra.bio/register/profile
({
  "causes": [
    {
      "causes": [],
      "message": "user rawls-qa@broad-dsde-qa.iam.gserviceaccount.com already exists",
      "source": "sam",
      "stackTrace": [],
      "statusCode": 409
    }
  ],
  "message": "user rawls-qa@broad-dsde-qa.iam.gserviceaccount.com already exists",
  "source": "Sam",
  "stackTrace": [],
  "statusCode": 409
})`)

	assert.Nil(t, _ignore409Conflict(err), "should suppress 409 Conflict errors")

	err = errors.New("a completely unexpected error")
	assert.Error(t, _ignore409Conflict(err), "should not ignore other errors")
}
