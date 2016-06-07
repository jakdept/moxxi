package moxxiConf

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"io/ioutil"
	"os"
	"text/template"
)


func LoadConfig() ([]string, []HandlerConfig, Err) {
	config, err := prepConfig()
	if err != nil {
		return []string{}, []HandlerConfig{}, err
	}

	if err = validateConfig(config); err != nil {
		return []string{}, []HandlerConfig{}, err
	}

	handlers, listens, err := loadConfig(config)
	if err != nil {
		return []string{}, []HandlerConfig{}, err
	}

	return handlers, listens, nil
}


func prepConfig() (*map[string]interface{}, Err) {

	possibleConfigs := []string{
		"./config.json",
		"/etc/moxxi/config.json",
		"$HOME/.moxxi/config.json",
		"./moxxi.config.json",
	}

	var data []byte
	var globErr error
	var c *map[string]interface{}
	for _, configTry := range possibleConfigs {
		data, globErr := ioutil.ReadFile(configTry)
		switch globErr {
		case os.ErrNotExist:
			continue
		case nil:
			break
		default:
			return nil, NewErr{
				Code: ErrConfigBadRead,
				fmt.Errorf("bad config file - %s", configTry),
			}
		}
	}

	globErr := json.Unmarshal(data, c)
	if globErr != nil {
		return nil, UpgradeError(globErr)
	}

	if _, ok := c["listen"]; !ok {
		c["listen"] = []string{"localhost:8080"}
	}

	return c
}

func validateConfig(c *map[string]interface{}) Err {
	// clean up array top level lines
	for _, part := range []string{"listen", "exclude"} {
		if _, ok := c[part]; ok {
			if _, ok := c[part].([]interface{}); !ok {
				c[part] = []interface{}{c[part]}
			}

			if _, ok := c[part].([]string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: part}
			}
		}
	}

	// check that the following things are strings in the top level
	for _, part := range []string{
		"baseURL",
		"confFile",
		"confExt",
		"confFile",
		"resFile",
	} {
		if _, ok := c[part]; ok {
			if _, ok := c[part].(string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: part}
			}
		}
	}

	if _, ok := c["subdomainLen"]; ok {
		var subdomainLen int
		if subdomainLen, ok := c["subdomainLen"].(int); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: "subdomainLen"}
		}
		if subdomainLen < 8 {
			c["subdomainLen"] = 8
		}
	}

	// verify the handlers are an array
	handlers, ok := c["handler"].([]interface{})
	if !ok {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handler",
		}
	}

	// test and propagate handlers
	for id, eachHandler := range handlers {
		locErr := validateConfigHandler(c, id, eachHandler)
		if locErr != nil {
			return locErr
		}
	}

	return nil
}

func loadConfig(c *viper.Viper) ([]string, []HandlerConfig, Err) {

	var handlers []HandlerConfig
	var listens []string

	err := c.UnmarshalKey("handler", &handlers)
	if err != nil {
		returnErr := NewErr{Code: ErrConfigBadExtract, value: "handlers", deepErr: err}
		return []string{}, []HandlerConfig{}, returnErr
	}

	err = c.UnmarshalKey("listen", &listens)
	if err != nil {
		returnErr := NewErr{Code: ErrConfigBadExtract, value: "listen", deepErr: err}
		return []string{}, []HandlerConfig{}, returnErr
	}

	return listens, handlers, nil
}
