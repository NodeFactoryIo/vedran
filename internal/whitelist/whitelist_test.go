package whitelist

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// helper functions

func createTmpWhitelistTestFile(t *testing.T) {
	file, _ := os.Create("./tmp_whitelist_test.txt")
	_, err := file.WriteString("node1\nnode2\nnode3")
	if err != nil {
		t.Fatal(err)
	}
}

func reset() {
	fileWithWhitelistedNodes = ""
	whitelistedNodes = nil
}

// tests

func Test_WhitelistingFromMemory(t *testing.T) {
	_ = InitWhitelistedNodes([]string{"node1", "node2", "node3"})

	tests := []struct {
		name string
		nodeID string
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
	createTmpWhitelistTestFile(t)
	defer os.Remove("./tmp_whitelist_test.txt")

	_ = InitWhitelistedNodesFromFile("./tmp_whitelist_test.txt")

	tests := []struct {
		name string
		nodeID string
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
	createTmpWhitelistTestFile(t)
	defer os.Remove("./tmp_whitelist_test.txt")

	assert.Nil(t, InitWhitelistedNodesFromFile("./tmp_whitelist_test.txt"))
	assert.Error(t, InitWhitelistedNodes([]string{"node1", "node2"}))

	reset()

	assert.Nil(t, InitWhitelistedNodes([]string{"node1", "node2"}))
	assert.Error(t, InitWhitelistedNodesFromFile("test-file.txt"))

	reset()
}

func Test_InitWhitelistedNodesFromFile_FileDoesntExist(t *testing.T) {
	assert.Error(t, InitWhitelistedNodesFromFile("test-file.txt"))
	reset()
}


