package moxxiConf

import (
	"encoding/json"
	"fmt"
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
	var c map[string]interface{}
	for _, configTry := range possibleConfigs {
		data, globErr = ioutil.ReadFile(configTry)
		switch globErr {
		case os.ErrNotExist:
			continue
		case nil:
			break
		default:
			return nil, NewErr{
				Code:    ErrConfigBadRead,
				deepErr: fmt.Errorf("bad config file - %s", configTry),
			}
		}
	}

	globErr = json.Unmarshal(data, &c)
	if globErr != nil {
		return nil, UpgradeError(globErr)
	}

	if _, ok := c["listen"]; !ok {
		c["listen"] = []string{"localhost:8080"}
	}

	return &c, nil
}

func validateConfig(dirtyConfig *map[string]interface{}) Err {
c := *dirtyConfig

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
		"confPath",
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
		if subdomainLen, ok = c["subdomainLen"].(int); !ok {
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
	for id, _ := range handlers {
		locErr := validateConfigHandler(&c, id)
		if locErr != nil {
			return locErr
		}
	}

	dirtyConfig = &c

	return nil
}

func validateConfigHandler(pConfig *map[string]interface{}, id int) Err {

	// unpack everything
	c := *pConfig
	var allHandlers []interface{}
	var h map[string]interface{}
	var ok bool

	if allHandlers, ok = c["handler"].([]interface{}); !ok {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handler",
		}
	}
	if h, ok = allHandlers[id].(map[string]interface{}); !ok {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handler",
		}
	}

	// check the exclude
	if _, ok := h["exclude"]; ok {
		// if it exists, make it an array
		if _, ok := h["exclude"].([]interface{}); !ok {
			h["exclude"] = []interface{}{h["exclude"]}
		}

		// check to make sure it's a string array
		if _, ok := h["exclude"].([]string); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: "exclude"}
		}
	} else {
		// if it does not exist, propagate it from above
		if _, ok := c["exclude"]; ok {
			h["exclude"] = c["exclude"]
		} else {
			h["exclude"] = []string{}
		}
	}

	// check the following things - validate and propagate them
	for _, part := range []string{
		"baseURL",
		"confPath",
		"confExt",
		"confFile",
		"resFile",
	} {
		if _, ok := h[part]; ok {
			if _, ok := h[part].(string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: part}
			}
		} else {
			if _, ok := c[part]; ok {
				h[part] = c[part]
			} else {
				h[part] = ""
			}
		}
	}

	if _, ok = h["subdomainLen"]; ok {
		var subdomainLen int
		if subdomainLen, ok = c["subdomainLen"].(int); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: "subdomainLen"}
		}
		if subdomainLen < 8 {
			c["subdomainLen"] = 8
		}
	} else {
		if _, ok := c["subdomainLen"]; ok {
			h["subdomainLen"] = c["subdomainLen"]
		} else {
			h["subdomainLen"] = 8
		}
	}

	// pack things back in
	allHandlers[id] = h
	c["handler"] = allHandlers
	pConfig = &c
	return nil
}

func loadConfig(pConfig *map[string]interface{}) (
	[]string, []HandlerConfig, Err) {

	c := *pConfig
	var ok bool
	var listens []string

	if listens, ok = c["listen"].([]string); !ok {
		return []string{}, []HandlerConfig{}, NewErr{
			Code:    ErrConfigBadStructure,
			value:   "listen",
			deepErr: fmt.Errorf("wrong type of %T", c["listen"]),
		}
	}

	var handlers []HandlerConfig
	var dirtyHandlers []interface{}

	if dirtyHandlers, ok = c["handler"].([]interface{}); !ok {
		return []string{}, []HandlerConfig{}, NewErr{
			Code:    ErrConfigBadStructure,
			value:   "handler",
			deepErr: fmt.Errorf("wrong type of %T", c["listen"]),
		}
	}

	for _, oneHandler := range dirtyHandlers {
		cleanHandler, err := decodeHandler(oneHandler)
		if err != nil {
			return []string{}, []HandlerConfig{}, err
		} else {
			handlers = append(handlers, cleanHandler)
		}
	}

	return listens, handlers, nil
}

func decodeHandler(dirtyHandler interface{}) (HandlerConfig, Err) {
	var h HandlerConfig

	var addressed map[string]interface{}
	var ok bool

	if addressed, ok = dirtyHandler.(map[string]interface{}); !ok {
		return HandlerConfig{}, NewErr{
			Code:    ErrConfigBadStructure,
			value:   fmt.Sprintf("%#v", dirtyHandler),
			deepErr: fmt.Errorf("bad handler in config"),
		}
	}

	if _, ok = addressed["handlerType"]; ok {
		if h.handlerType, ok = addressed["handlerType"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "handlerType",
			}
		}
	}
	if _, ok = addressed["handlerRoute"]; ok {
		if h.handlerRoute, ok = addressed["handlerRoute"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "handlerRoute",
			}
		}
	}

	if _, ok = addressed["baseURL"]; ok {
		if h.baseURL, ok = addressed["baseURL"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "baseURl",
			}
		}
	}
	if _, ok = addressed["confPath"]; ok {
		if h.confPath, ok = addressed["confPath"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "confPath",
			}
		}
	}
	if _, ok = addressed["confExt"]; ok {
		if h.confExt, ok = addressed["confExt"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "confExt",
			}
		}
	}

	if _, ok = addressed["exclude"]; ok {
		if h.exclude, ok = addressed["exclude"].([]string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "exclude",
			}
		}
	}

	if _, ok = addressed["subdomainLen"]; ok {
		if h.subdomainLen, ok = addressed["subdomainLen"].(int); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "subdomainLen",
			}
		}
	}

	var err error
	if _, ok = addressed["confFile"]; ok {
		if workFile, ok := addressed["confFile"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "confFile",
			}
		} else {
			h.confTempl, err = template.ParseFiles(workFile)
			if err != nil {
				return HandlerConfig{}, UpgradeError(err)
			}
		}
	}
	if _, ok = addressed["resFile"]; ok {
		if workFile, ok := addressed["resFile"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigBadStructure,
				value: "resFile",
			}
		} else if addressed["handlerType"] != "static" {
			h.resTempl, err = template.ParseFiles(workFile)
			if err != nil {
				return HandlerConfig{}, UpgradeError(err)
			}
		} else {
			h.resFile = workFile
		}
	}

	return h, nil
}
