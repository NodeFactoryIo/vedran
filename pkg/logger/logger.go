package logger

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// SetupLogger sets logs and leves and initializes logger
func SetupLogger(level log.Level, logfilePath string) error {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(level)

	if logfilePath != "" {
		file, err := os.OpenFile(logfilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}

	return nil
}
