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

func IsNodeWhitelisted(nodeId string) bool {
	if fileWithWhitelistedNodes != "" {
		// read whitelisted from file
		file, err := ioutil.ReadFile(fileWithWhitelistedNodes)
		if err != nil {
			log.Errorf("Unable to read file with whitelisted nodes %s because %v", fileWithWhitelistedNodes, err)
			return false
		}
		for _, nodeIdBytes := range bytes.Split(file, []byte{'\n'}) {
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

func InitWhitelistedNodes(nodes []string) error {
	if whitelistedNodes == nil && fileWithWhitelistedNodes == "" {
		whitelistedNodes = nodes
		return nil
	} else {
		return errors.New("whitelisted nodes already initialized")
	}
}

func InitWhitelistedNodesFromFile(filePath string) error {
	if fileWithWhitelistedNodes == "" && whitelistedNodes == nil {
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			return errors.New(fmt.Sprintf("whitelisted nodes file %s doesn't exist", filePath))
		}
		fileWithWhitelistedNodes = filePath
		return nil
	} else {
		return errors.New("whitelisted nodes already initialized")
	}
}