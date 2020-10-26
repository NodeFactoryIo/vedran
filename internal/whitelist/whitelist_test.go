package whitelist

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

// helper functions

func createTmpWhitelistTestFile(t *testing.T, fileContent string) {
	file, _ := os.Create("./tmp_whitelist_test.txt")
	_, err := file.WriteString(fileContent)
	if err != nil {
		t.Fatal(err)
	}
}

func reset() {
	fileWithWhitelistedNodes = ""
	whitelistedNodes = nil
}

// tests

func Test_RemoveNodeFromWhitelisted_FromMemory(t *testing.T) {
	tests := []struct {
		name                        string
		whitelistedNodes            []string
		removeNodeId                string
		removeFails                 bool
		whitelistedNodesAfterRemove []string
	}{
		{
			name:                        "remove one whitelisted node",
			whitelistedNodes:            []string{"node1", "node2", "node3"},
			removeNodeId:                "node2",
			removeFails:                 false,
			whitelistedNodesAfterRemove: []string{"node1", "node3"},
		},
		{
			name:                        "fail on removing not whitelisted node",
			whitelistedNodes:            []string{"node1", "node2", "node3"},
			removeNodeId:                "node4",
			removeFails:                 true,
			whitelistedNodesAfterRemove: []string{"node1", "node3"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_ = initWhitelistedNodes(test.whitelistedNodes)
			err := RemoveNodeFromWhitelisted(test.removeNodeId)
			if test.removeFails {
				assert.Error(t, err)
				assert.Equal(t, test.whitelistedNodes, whitelistedNodes)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.whitelistedNodesAfterRemove, whitelistedNodes)
			}
			reset()
		})
	}
}

func Test_RemoveNodeFromWhitelisted_FromFile(t *testing.T) {
	tests := []struct {
		name                       string
		removeNodeId               string
		removeFails                bool
		whitelistedFile            string
		whitelistedFileAfterRemove string
	}{
		{
			name:                       "remove one whitelisted node",
			whitelistedFile:            "node1\nnode2\nnode3",
			whitelistedFileAfterRemove: "node1\nnode3",
			removeNodeId:               "node2",
			removeFails:                false,
		},
		{
			name:                       "remove one whitelisted node",
			whitelistedFile:            "node1\nnode2\nnode3",
			whitelistedFileAfterRemove: "node1\nnode2\nnode3",
			removeNodeId:               "node4",
			removeFails:                false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			createTmpWhitelistTestFile(t, test.whitelistedFile)
			defer os.Remove("./tmp_whitelist_test.txt")

			_ = initWhitelistedNodesFromFile("./tmp_whitelist_test.txt")

			err := RemoveNodeFromWhitelisted(test.removeNodeId)

			fileContent, _ := ioutil.ReadFile("./tmp_whitelist_test.txt")
			fileContentString := string(fileContent)
			if test.removeFails {
				assert.Error(t, err)
				assert.Equal(t, test.whitelistedFile, fileContentString)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.whitelistedFileAfterRemove, fileContentString)
			}
			reset()
		})
	}
}

func Test_WhitelistingFromMemory(t *testing.T) {
	_ = initWhitelistedNodes([]string{"node1", "node2", "node3"})

	tests := []struct {
		name     string
		nodeID   string
		expected bool
	}{
		{name: "whitelisted node1", nodeID: "node1", expected: true},
		{name: "whitelisted node2", nodeID: "node2", expected: true},
		{name: "whitelisted node3", nodeID: "node3", expected: true},
		{name: "not whitelisted node4", nodeID: "node4", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsNodeWhitelisted(test.nodeID))
		})
	}

	reset()
}

func Test_WhitelistingFromFile(t *testing.T) {
	createTmpWhitelistTestFile(t, "node1\nnode2\nnode3")
	defer os.Remove("./tmp_whitelist_test.txt")

	_ = initWhitelistedNodesFromFile("./tmp_whitelist_test.txt")

	tests := []struct {
		name     string
		nodeID   string
		expected bool
	}{
		{name: "whitelisted node1", nodeID: "node1", expected: true},
		{name: "whitelisted node2", nodeID: "node2", expected: true},
		{name: "whitelisted node3", nodeID: "node3", expected: true},
		{name: "not whitelisted node4", nodeID: "node4", expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, IsNodeWhitelisted(test.nodeID))
		})
	}

	reset()
}

func Test_WhitelistInit(t *testing.T) {
	createTmpWhitelistTestFile(t, "node1\nnode2\nnode3")
	defer os.Remove("./tmp_whitelist_test.txt")

	assert.Nil(t, initWhitelistedNodesFromFile("./tmp_whitelist_test.txt"))
	assert.Error(t, initWhitelistedNodes([]string{"node1", "node2"}))

	reset()

	assert.Nil(t, initWhitelistedNodes([]string{"node1", "node2"}))
	assert.Error(t, initWhitelistedNodesFromFile("test-file.txt"))

	reset()
}

func Test_InitWhitelistedNodesFromFile_FileDoesntExist(t *testing.T) {
	assert.Error(t, initWhitelistedNodesFromFile("test-file.txt"))
	reset()
}
