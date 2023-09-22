package render

import (
	"github.com/broadinstitute/thelma/internal/thelma/app"
	"github.com/broadinstitute/thelma/internal/thelma/render"
	"github.com/broadinstitute/thelma/internal/thelma/render/helmfile"
)

// renderWrapper a simple wrapper around render.DoRender that allows to intercept calls for tests
// if I rewrote this code today the render package would have a mockable interface but this is a quick fix
type renderWrapper interface {
	// doRender Delegates to render.DoRender
	doRender(app app.ThelmaApp, globalOptions *render.Options, helmfileArgs *helmfile.Args) error
}

func newRenderWrapper() renderWrapper {
	return realWrapper{}
}

type realWrapper struct{}

func (r realWrapper) doRender(app app.ThelmaApp, globalOptions *render.Options, helmfileArgs *helmfile.Args) error {
	return render.DoRender(app, globalOptions, helmfileArgs)
}
