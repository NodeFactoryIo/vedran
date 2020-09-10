package config

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/spf13/viper"
)

func TestInitMainConfig_getMainConfigName(t *testing.T) {
	_ = os.Setenv("ENV", "test")
	configName := getMainConfigName()
	assert.Equal(t, configName, "config-test")

	_ = os.Setenv("ENV", "")
	configName = getMainConfigName()
	assert.Equal(t, configName, "config")
}

func TestInitMainConfig_testDefaultValues(t *testing.T) {
	InitMainConfig()
	assert.Equal(t, viper.GetInt("stats.interval"), 30)
	assert.Equal(t, viper.GetString("log.level"), "error")
}
