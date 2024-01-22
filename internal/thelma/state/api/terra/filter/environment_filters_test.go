package filter

import (
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra"
	"github.com/broadinstitute/thelma/internal/thelma/state/api/terra/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEnvironmentFilters(t *testing.T) {
	noAutoDelete := &mocks.AutoDelete{}
	noAutoDelete.EXPECT().Enabled().Return(false)

	autoDeleteAfter2HoursAgo := &mocks.AutoDelete{}
	autoDeleteAfter2HoursAgo.EXPECT().Enabled().Return(true)
	autoDeleteAfter2HoursAgo.EXPECT().After().Return(time.Now().Add(-1 * 2 * time.Hour))

	dev := &mocks.Environment{}
	dev.EXPECT().Name().Return("dev")
	dev.EXPECT().Lifecycle().Return(terra.Static)
	dev.EXPECT().Base().Return("live")
	dev.EXPECT().Template().Return("")
	dev.EXPECT().CreatedAt().Return(time.Now().Add(-1 * 24 * 100 * time.Hour)) // 100 days old
	dev.EXPECT().AutoDelete().Return(noAutoDelete)
	dev.EXPECT().PreventDeletion().Return(true)

	swat := &mocks.Environment{}
	swat.EXPECT().Name().Return("swatomation")
	swat.EXPECT().Lifecycle().Return(terra.Template)
	swat.EXPECT().Base().Return("bee")
	swat.EXPECT().Template().Return("")
	swat.EXPECT().CreatedAt().Return(time.Now().Add(-1 * 24 * 30 * time.Hour)) // 30 days old
	swat.EXPECT().AutoDelete().Return(noAutoDelete)
	swat.EXPECT().PreventDeletion().Return(false)

	bee := &mocks.Environment{}
	bee.EXPECT().Name().Return("my-bee")
	bee.EXPECT().Lifecycle().Return(terra.Dynamic)
	bee.EXPECT().Base().Return("bee")
	bee.EXPECT().Template().Return("swatomation")
	bee.EXPECT().CreatedAt().Return(time.Now().Add(-1 * 6 * time.Hour)) // 6 hours old
	bee.EXPECT().AutoDelete().Return(autoDeleteAfter2HoursAgo)
	bee.EXPECT().PreventDeletion().Return(false)

	testCases := []struct {
		filter terra.EnvironmentFilter
		expect []terra.Environment
	}{
		{
			filter: Environments().Any(),
			expect: []terra.Environment{dev, swat, bee},
		},
		{
			filter: Environments().Any().Negate(),
			expect: nil,
		},
		{
			filter: Environments().IsTemplate(),
			expect: []terra.Environment{swat},
		},
		{
			filter: Environments().IsTemplate().Negate(),
			expect: []terra.Environment{dev, bee},
		},
		{
			filter: Environments().HasBase("live"),
			expect: []terra.Environment{dev},
		},
		{
			filter: Environments().HasBase("live").Negate(),
			expect: []terra.Environment{swat, bee},
		},
		{
			filter: Environments().HasBase("bee"),
			expect: []terra.Environment{swat, bee},
		},
		{
			filter: Environments().HasBase("bee").Negate(),
			expect: []terra.Environment{dev},
		},
		{
			filter: Environments().HasBase("live", "bee").Negate(),
			expect: nil,
		},
		{
			filter: Environments().HasLifecycle(terra.Template),
			expect: []terra.Environment{swat},
		},
		{
			filter: Environments().HasLifecycle(terra.Dynamic),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().HasLifecycle(terra.Static),
			expect: []terra.Environment{dev},
		},
		{
			filter: Environments().HasLifecycleName("template"),
			expect: []terra.Environment{swat},
		},
		{
			filter: Environments().HasLifecycleName("dynamic"),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().HasLifecycleName("static"),
			expect: []terra.Environment{dev},
		},
		{
			filter: Environments().HasLifecycleName("static", "template"),
			expect: []terra.Environment{dev, swat},
		},
		{
			filter: Environments().HasTemplate(swat),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().HasTemplateName("swatomation"),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().NameIncludes("e"),
			expect: []terra.Environment{dev, bee},
		},
		{
			filter: Environments().OlderThan(12 * time.Hour),
			expect: []terra.Environment{dev, swat},
		},
		{
			filter: Environments().OlderThan(60 * 24 * time.Hour),
			expect: []terra.Environment{dev},
		},
		{
			filter: Environments().OlderThan(4 * time.Hour).And(Environments().HasLifecycle(terra.Dynamic)),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().OlderThan(24 * time.Hour).Or(Environments().HasLifecycle(terra.Dynamic)),
			expect: []terra.Environment{bee, dev, swat},
		},
		{
			filter: Environments().AutoDeletable(),
			expect: []terra.Environment{bee},
		},
		{
			filter: Environments().AutoDeletable().Negate(),
			expect: []terra.Environment{dev, swat},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filter.String(), func(t *testing.T) {
			assert.ElementsMatch(t, tc.expect, tc.filter.Filter([]terra.Environment{dev, swat, bee}))

		})
	}
}
