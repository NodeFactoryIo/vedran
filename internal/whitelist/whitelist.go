package whitelist

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
)

var (
	fileWithWhitelistedNodes string
	whitelistedNodes []string
)

var newLine = []byte{'\n'}

func InitWhitelisting(whitelistedNodes []string, whitelistFile string) (bool, error) {
	var whitelistError error
	whitelistEnabled := true
	if whitelistFile != "" {
		whitelistError = initWhitelistedNodesFromFile(whitelistFile)
	} else if len(whitelistedNodes) != 0 {
		whitelistError = initWhitelistedNodes(whitelistedNodes)
	} else {
		whitelistEnabled = false
	}
	return whitelistEnabled, whitelistError
}

func initWhitelistedNodes(nodes []string) error {
	if whitelistedNodes == nil && fileWithWhitelistedNodes == "" {
		whitelistedNodes = nodes
		return nil
	} else {
		return errors.New("whitelisted nodes already initialized")
	}
}

func initWhitelistedNodesFromFile(filePath string) error {
	if fileWithWhitelistedNodes == "" && whitelistedNodes == nil {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			return fmt.Errorf("whitelisted nodes file %s doesn't exist", filePath)
		}
		fileWithWhitelistedNodes = filePath
		return nil
	} else {
		return errors.New("whitelisted nodes already initialized")
	}
}

func IsNodeWhitelisted(nodeId string) bool {
	if fileWithWhitelistedNodes != "" {
		file, err := ioutil.ReadFile(fileWithWhitelistedNodes)
		if err != nil {
			log.Errorf("Unable to read file with whitelisted nodes %s because %v", fileWithWhitelistedNodes, err)
			return false
		}
		for _, nodeIdBytes := range bytes.Split(file, newLine) {
			if string(nodeIdBytes) == nodeId {
				return true
			}
		}
		return false
	} else {
		for _, wn := range whitelistedNodes {
			if wn == nodeId {
				return true
			}
		}
		return false
	}
}

func RemoveNodeFromWhitelisted(nodeId string) error {
	if fileWithWhitelistedNodes != "" {
		return removeNodeFromWhitelistFile(nodeId)
	} else if len(whitelistedNodes) != 0 {
		return removeNodeFromWhitelistArray(nodeId)
	} else {
		// whitelisting disabled
		return nil
	}
}

func removeNodeFromWhitelistFile(nodeId string) error {
	file, err := ioutil.ReadFile(fileWithWhitelistedNodes)
	if err != nil {
		return err
	}
	// remove node id from file content
	var newFileContent []byte
	lines := bytes.Split(file, newLine)
	for i, nodeIdBytes := range lines {
		if string(nodeIdBytes) != nodeId {
			newFileContent = append(newFileContent, nodeIdBytes...)
			if i + 1 < len(lines) { // append new line expect last entry
				newFileContent = append(newFileContent, newLine...)
			}
		}
	}
	// save new file content to file
	err = ioutil.WriteFile(fileWithWhitelistedNodes, newFileContent, 0644)
	if err != nil {
		return err
	}
	return nil
}

func removeNodeFromWhitelistArray(nodeId string) error {
	// find node id inside array
	removeIndex := -1
	for i, id := range whitelistedNodes {
		if id == nodeId {
			removeIndex = i
		}
	}
	// remove node if index found
	if removeIndex != -1 {
		whitelistedNodes[removeIndex] = whitelistedNodes[len(whitelistedNodes) - 1]
		whitelistedNodes[len(whitelistedNodes) - 1] = ""
		whitelistedNodes = whitelistedNodes[:len(whitelistedNodes) - 1]
	} else {
		return fmt.Errorf("node %s not found in whitelisted nodes", nodeId)
	}
	return nil
}
