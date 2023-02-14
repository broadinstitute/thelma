package toolbox

import (
	"os"
	"path"
)

type Toolbox interface {
	// ExpandPath if `command` is the name of a tool that exists in the toolbox directory,
	// such as "kubectl", "helm", "argocd", return the fully-qualified path to that tool
	// within the toolbox directory.
	// (eg. "/Users/blah/.thelma/releases/v1.2.3/tools/bin/helm")
	// else, return the name unchanged (eg. "curl")
	ExpandPath(command string) string
}

type toolbox struct {
	dir       string
	toolNames map[string]struct{}
}

func New(dir string) (Toolbox, error) {
	names := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names[entry.Name()] = struct{}{}
	}
	return &toolbox{
		dir:       dir,
		toolNames: names,
	}, nil
}

func (t *toolbox) ExpandPath(name string) string {
	if _, exists := t.toolNames[name]; exists {
		return path.Join(t.dir, name)
	}
	return name
}
