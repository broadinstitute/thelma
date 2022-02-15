package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"sort"
	"strings"
)

// node represents a node in the command tree
type node struct {
	key           commandKey
	thelmaCommand ThelmaCommand
	cobraCommand  *cobra.Command
	children      map[string]*node
	parent        *node
}

// newTree constructs a new command tree from a map of Thelma commands, keyed by name
func newTree(commands map[string]ThelmaCommand) *node {
	// bundle commands with their keys together in a slice
	type entry struct {
		key     commandKey
		command ThelmaCommand
	}

	var entries []entry

	for fullName, command := range commands {
		entries = append(entries, entry{
			key:     newCommandKey(fullName),
			command: command,
		})
	}

	// sort slice by command depth, so that shorter/intermediate commands come before longer
	// eg.
	// "charts", "render", "version", "charts import", "charts publish"
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].key.depth() < entries[j].key.depth()
	})

	// build tree
	// create root command
	root := newNode(rootCommandKey(), newRootCommand(), nil)

	// build command tree
	for _, entry := range entries {
		// find parent
		parent := findNode(root, entry.key.ancestors())
		if parent == nil {
			cmdNames := make([]string, len(entries))
			for i, k := range entries {
				cmdNames[i] = fmt.Sprintf("%q", k.key.description())
			}
			panic(fmt.Errorf("could not find parent command for command %s, registered command names are:\n%v", entry.key, strings.Join(cmdNames, "\n")))
		}

		// add a node for this entry
		_node := newNode(entry.key, entry.command, parent)
		parent.addChild(_node)
	}

	return root
}

// constructor for node
func newNode(key commandKey, thelmaCommand ThelmaCommand, parent *node) *node {
	return &node{
		key:           key,
		thelmaCommand: thelmaCommand,
		cobraCommand:  &cobra.Command{},
		children:      make(map[string]*node),
		parent:        parent,
	}
}

// true if this node is the root
func (n *node) isRoot() bool {
	return n.parent == nil
}

// true if this node is a leaf
func (n *node) isLeaf() bool {
	return len(n.children) == 0
}

// adds a child to the node
func (n *node) addChild(child *node) {
	n.children[child.key.shortName()] = child
}

// find a node, given a command anme like ["charts", "import"]
func findNode(root *node, fullName []string) *node {
	if root == nil {
		return nil
	}
	parent := root
	for _, component := range fullName {
		current, exists := parent.children[component]
		if !exists {
			return nil
		}
		parent = current
	}
	return parent
}

// walk tree, invoking callback on every node
func preOrderTraverse(n *node, callback func(*node)) {
	callback(n)
	for _, child := range n.children {
		preOrderTraverse(child, callback)
	}
}

// returns all nodes on the path from n to root, including n.
// the root node will be the last item in the slice
func pathToRoot(n *node) []*node {
	var result []*node
	for n != nil {
		result = append(result, n)
		n = n.parent
	}
	return result
}

// returns all nodes on the path the root to n, including n.
// the root node will be the first item in the slice
func pathFromRoot(n *node) []*node {
	path := pathToRoot(n)
	for i := 0; i < len(path)/2; i++ {
		j := len(path) - i - 1
		path[i], path[j] = path[j], path[i]
	}
	return path
}
