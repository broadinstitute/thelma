package toolbox

import (
	"os"
	"path"
)

// ToolFinder exists to help shell.Runner instances resolve fully-qualified paths to Thelma's bundled tools
type ToolFinder interface {
	// ExpandPath if `command` is the name of a tool that exists in the toolbox directory,
	// such as "kubectl", "helm", "argocd", return the fully-qualified path to that tool
	// within the toolbox directory.
	// (eg. "/Users/blah/.thelma/releases/v1.2.3/tools/bin/helm")
	// else, return the name unchanged (eg. "curl")
	ExpandPath(command string) string
}

type finder struct {
	executableDir string
	toolNames     map[string]struct{}
}

// NewToolFinder calls FindToolsDir to identify the location of Thelma's bundled tools
// and returns a new ToolFinder rooted at the tools dir's `bin` directory
func NewToolFinder() (ToolFinder, error) {
	toolsRoot, err := FindToolsDir()
	if err != nil {
		return nil, err
	}

	exedir := path.Join(toolsRoot, executableDirName)

	return NewToolFinderWithDir(exedir)
}

// NewToolFinderWithDir FOR USE IN TESTS ONLY construct a new ToolFinder rooted at a custom tools directory
func NewToolFinderWithDir(executableDir string) (ToolFinder, error) {
	names := make(map[string]struct{})
	entries, err := os.ReadDir(executableDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names[entry.Name()] = struct{}{}
	}
	return &finder{
		executableDir: executableDir,
		toolNames:     names,
	}, nil
}

func (t *finder) ExpandPath(name string) string {
	if _, exists := t.toolNames[name]; exists {
		return path.Join(t.executableDir, name)
	}
	return name
}
