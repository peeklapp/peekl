package inventory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// openRoot opens dir as an *os.Root and closes it at the end of the test.
func openRoot(t *testing.T, dir string) *os.Root {
	t.Helper()
	root, err := os.OpenRoot(dir)
	require.NoError(t, err, "opening root")
	return root
}

// writeCode creates a temporary code root populated with the given files
// (path relative to the root -> content) and returns it opened.
func writeCode(t *testing.T, files map[string]string) *os.Root {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return openRoot(t, dir)
}

func TestLoadInventory(t *testing.T) {
	nodes, groups, err := LoadInventory(openRoot(t, "testdata/valid_inventory"))
	require.NoError(t, err)

	t.Run("nodes", func(t *testing.T) {
		// README.md in the nodes directory must be ignored.
		require.Len(t, nodes, 2)

		dummy, ok := nodes["dummy"]
		require.True(t, ok, "node 'dummy' should be loaded")
		assert.Equal(t, "dummy", dummy.Name)
		assert.Equal(t, "inventory/nodes/dummy.hcl", dummy.SourcePath)
		assert.Equal(t, []string{"web"}, dummy.Groups)
		assert.Equal(t, []string{"tag_from_node"}, dummy.Tags)
		assert.Equal(t, map[string]any{
			"environment": "production",
			"replicas":    3,
		}, dummy.Variables)

		require.Len(t, dummy.Roles, 1)
		assert.Equal(t, "test", dummy.Roles[0].Name)
		assert.Equal(t, map[string]any{}, dummy.Roles[0].Variables)
	})

	t.Run("node with only a name", func(t *testing.T) {
		other, ok := nodes["other"]
		require.True(t, ok, "node 'other' should be loaded")
		assert.Equal(t, "inventory/nodes/other.hcl", other.SourcePath)
		assert.Empty(t, other.Groups)
		assert.Empty(t, other.Tags)
		assert.Empty(t, other.Roles)
		assert.NotNil(t, other.Variables, "variables should default to an empty map")
		assert.Empty(t, other.Variables)
	})

	t.Run("groups", func(t *testing.T) {
		require.Len(t, groups, 1)

		web, ok := groups["web"]
		require.True(t, ok, "group 'web' should be loaded")
		assert.Equal(t, "web", web.Name)
		assert.Equal(t, "inventory/groups/web.hcl", web.SourcePath)
		assert.Equal(t, []string{"tag_from_group"}, web.Tags)
		assert.Equal(t, map[string]any{"listen_port": 80}, web.Variables)

		require.Len(t, web.Roles, 1)
		assert.Equal(t, "nginx", web.Roles[0].Name)
		assert.Equal(t, map[string]any{
			"worker_processes": 4,
			"server_names":     []any{"example.com", "www.example.com"},
			"gzip":             true,
		}, web.Roles[0].Variables)
	})
}

func TestLoadInventoryEmpty(t *testing.T) {
	cases := map[string]map[string]string{
		"no inventory directory": {},
		"empty inventory directories": {
			"inventory/nodes/.keep":  "",
			"inventory/groups/.keep": "",
		},
	}

	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			nodes, groups, err := LoadInventory(writeCode(t, files))
			require.NoError(t, err)
			assert.Empty(t, nodes)
			assert.Empty(t, groups)
		})
	}
}

func TestLoadInventoryErrors(t *testing.T) {
	cases := []struct {
		name        string
		files       map[string]string
		errContains string
	}{
		{
			name: "node name does not match file name",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "other" {}`,
			},
			errContains: "node 'other' found in file 'inventory/nodes/dummy.hcl'. The file should be named 'inventory/nodes/other.hcl'",
		},
		{
			name: "group name does not match file name",
			files: map[string]string{
				"inventory/groups/web.hcl": `group "db" {}`,
			},
			errContains: "group 'db' found in file 'inventory/groups/web.hcl'. The file should be named 'inventory/groups/db.hcl'",
		},
		{
			name: "invalid node HCL syntax",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" {`,
			},
			errContains: "inventory/nodes/dummy.hcl",
		},
		{
			name: "invalid group HCL syntax",
			files: map[string]string{
				"inventory/groups/web.hcl": `group "web" {`,
			},
			errContains: "inventory/groups/web.hcl",
		},
		{
			name: "node file without node block",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `group "dummy" {}`,
			},
			errContains: "inventory/nodes/dummy.hcl",
		},
		{
			name: "node block without label",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node {}`,
			},
			errContains: "inventory/nodes/dummy.hcl",
		},
		{
			name: "unknown attribute in node",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" { unknown = true }`,
			},
			errContains: "unknown",
		},
		{
			name: "node groups is not a list",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" { groups = 42 }`,
			},
			errContains: "inventory/nodes/dummy.hcl",
		},
		{
			name: "node variables is not an object",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" { variables = "nope" }`,
			},
			errContains: "invalid variables for node 'dummy'",
		},
		{
			name: "group variables is not an object",
			files: map[string]string{
				"inventory/groups/web.hcl": `group "web" { variables = ["a", "b"] }`,
			},
			errContains: "invalid variables for group 'web'",
		},
		{
			name: "node role variables is not an object",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" {
  role "nginx" {
    variables = 1
  }
}`,
			},
			errContains: "invalid variables for role 'nginx'",
		},
		{
			name: "group role variables is not an object",
			files: map[string]string{
				"inventory/groups/web.hcl": `group "web" {
  role "nginx" {
    variables = true
  }
}`,
			},
			errContains: "invalid variables for role 'nginx'",
		},
		{
			name: "invalid group after valid node",
			files: map[string]string{
				"inventory/nodes/dummy.hcl": `node "dummy" {}`,
				"inventory/groups/web.hcl":  `group "db" {}`,
			},
			errContains: "group 'db' found in file 'inventory/groups/web.hcl'",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := LoadInventory(writeCode(t, tc.files))
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.errContains)
		})
	}
}

func TestLoadInventoryIgnoresNestedFiles(t *testing.T) {
	nodes, groups, err := LoadInventory(writeCode(t, map[string]string{
		"inventory/nodes/dummy.hcl":          `node "dummy" {}`,
		"inventory/nodes/archive/old.hcl":    `node "old" {}`,
		"inventory/groups/nested/broken.hcl": `group "broken" {`,
		"inventory/groups/web.hcl.disabled":  `group "web" {}`,
		"inventory/dummy.hcl":                `node "dummy" {}`,
	}))
	require.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Contains(t, nodes, "dummy")
	assert.Empty(t, groups)
}
