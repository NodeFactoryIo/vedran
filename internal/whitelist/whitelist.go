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