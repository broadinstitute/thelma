package github

import (
	"context"
	"testing"

	"github.com/google/go-github/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/stretchr/testify/suite"
)

func Test_GithubClientSuite(t *testing.T) {
	suite.Run(t, new(githubClientSuite))
}

type githubClientSuite struct {
	suite.Suite
}

func (suite *githubClientSuite) TestCallingUser() {
	mockClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetUser,
			github.User{
				Login: github.String("test-user"),
			},
		),
	)

	gh, err := New(WithClient(mockClient))
	suite.Assert().NoError(err)

	user, err := gh.GetCallingUser(context.Background())
	suite.Assert().NoError(err)
	suite.Assert().Equal("test-user", user)
}
