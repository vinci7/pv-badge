package conf

import (
	"os"
	"io/ioutil"
	"errors"

	"github.com/BurntSushi/toml"
	"github.com/labstack/gommon/log"
)

var (
	Conf              config // holds the global app config.
	defaultConfigFile = "conf/conf.toml"
)

type config struct {
	XLcId 		  string	`toml:"XLcId"`
	XLcKey 		  string	`toml:"XLcKey"`
	ContentType   string	`toml:"ContentType"`
	TotalPvUrl	  string	`toml:"TotalPvUrl"`
	TodayPvUrl	  string	`toml:"TodayPvUrl"`
}


// initConfig initializes the app configuration by first setting defaults,
// then overriding settings from the app config file, then overriding
// It returns an error if any.
func InitConfig() error {
	configFile := defaultConfigFile

	// Set defaults.
	Conf = config{}

	if _, err := os.Stat(configFile); err != nil {
		return errors.New("config file err:" + err.Error())
	} else {
		log.Infof("load config from file:" + configFile)
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return errors.New("config load err:" + err.Error())
		}
		_, err = toml.Decode(string(configBytes), &Conf)
		if err != nil {
			return errors.New("config decode err:" + err.Error())
		}
	}

	log.Infof("config data:%v", Conf)

	return nil
}