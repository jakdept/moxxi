package moxxiConf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

func LoadConfig() ([]string, string, string, []HandlerConfig, Err) {
	config, err := prepConfig()
	if err != nil {
		return nil, "", "", []HandlerConfig{}, err
	}

	if err = validateConfig(config); err != nil {
		return nil, "", "", []HandlerConfig{}, err
	}

	listens, accessLog, errorLog, handlers, err := loadConfig(config)
	if err != nil {
		return nil, "", "", []HandlerConfig{}, err
	}

	return listens, accessLog, errorLog, handlers, nil
}

func prepConfig() (*map[string]interface{}, Err) {

	possibleConfigs := []string{
		"./config.json",
		"/etc/moxxi/config.json",
		"$HOME/.moxxi/config.json",
		"./moxxi.config.json",
		"./test.config.json",
	}

	var data []byte
	var globErr error
	var c map[string]interface{}
ValidConfig:
	for _, configTry := range possibleConfigs {
		data, globErr = ioutil.ReadFile(configTry)
		switch {
		case os.IsNotExist(globErr):
		case globErr == nil:
			break ValidConfig
		default:
			return nil, NewErr{
				Code:    ErrConfigBadRead,
				value:   configTry,
				deepErr: globErr,
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
			chkArray, ok := c[part].([]interface{})
			if !ok {
				c[part] = []interface{}{c[part]}
				chkArray = []interface{}{c[part]}
			}

			// cannot type assert to ([]string) so i have to go through the parts?
			for _, each := range chkArray {
				if _, ok = each.(string); !ok {
					return NewErr{
						Code:    ErrConfigBadStructure,
						value:   part,
						deepErr: fmt.Errorf("%T - %#v", c[part], c[part]),
					}
				}
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
		"ipFile",
		"accessLog",
		"errorLog",
	} {
		if _, ok := c[part]; ok {
			if _, ok := c[part].(string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: part}
			}
		}
	}

	if _, ok := c["subdomainLen"]; ok {
		switch c["subdomainLen"].(type) {
		case int:
		case float64:
			c["subdomainLen"] = int(c["subdomainLen"].(float64))
		default:
			return NewErr{
				Code:  ErrConfigBadStructure,
				value: "subdomainLen",
				deepErr: fmt.Errorf("handler sub portion wrong type %T - %#v",
					c["subdomainLen"], c["subdomainLen"]),
			}
		}
		if c["subdomainLen"].(int) < 8 {
			c["subdomainLen"] = 8
		}
	}

	if _, ok := c["redirectTracing"]; ok {
		if _, ok = c["redirectTracing"].(bool); !ok {
			return NewErr{
				Code:    ErrConfigBadStructure,
				value:   "redirectTracing",
				deepErr: fmt.Errorf("%T - %#v", c["redirectTracing"], c["redirectTracing"]),
			}
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

	switch h["handlerType"] {
	case "static":
	case "form":
	case "json":
	default:
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handlerType",
		}
	}

	if loc, ok := h["handlerRoute"].(string); !ok {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handlerRoute",
		}
	} else if !strings.HasSuffix(loc, "/") {
		h["handlerRoute"] = loc + "/"
	}

	// check the exclude
	if _, ok := h["exclude"]; ok {
		// if it exists, make it an array
		if _, ok := h["exclude"].([]interface{}); !ok {
			h["exclude"] = []interface{}{h["exclude"]}
		}

		// cannot type assert to ([]string) so have to go through the parts
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
		"ipFile",
	} {
		if _, ok := h[part]; ok {
			if _, ok := h[part].(string); !ok {
				return NewErr{
					Code:    ErrConfigBadStructure,
					value:   part,
					deepErr: fmt.Errorf("handler portion - %T - %#v", h[part], h[part]),
				}
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
		switch h["subdomainLen"].(type) {
		case int:
		case float64:
			h["subdomainLen"] = int(h["subdomainLen"].(float64))
		default:
			return NewErr{
				Code:  ErrConfigBadStructure,
				value: "subdomainLen",
				deepErr: fmt.Errorf("handler sub portion wrong type %T - %#v",
					h["subdomainLen"], h["subdomainLen"]),
			}
		}
		if h["subdomainLen"].(int) < 8 {
			h["subdomainLen"] = 8
		}
	} else {
		if _, ok := c["subdomainLen"]; ok {
			h["subdomainLen"] = c["subdomainLen"]
		} else {
			h["subdomainLen"] = 8
		}
	}

	if _, ok = h["redirectTracing"]; ok {
		if _, ok = h["redirectTracing"].(bool); !ok {
			return NewErr{
				Code:    ErrConfigBadStructure,
				value:   "redirectTracing",
				deepErr: fmt.Errorf("%T - %#v", c["redirectTracing"], c["redirectTracing"]),
			}
		}
	} else {
		if _, ok := c["redirectTracing"]; ok {
			h["redirectTracing"] = c["redirectTracing"]
		}
	}

	// pack things back in
	allHandlers[id] = h
	c["handler"] = allHandlers
	pConfig = &c
	return nil
}

func loadConfig(pConfig *map[string]interface{}) (
	[]string, string, string, []HandlerConfig, Err) {

	c := *pConfig
	var ok bool
	var listens []string
	var accessLog, errorLog string

	if untypedListens, ok := c["listen"].([]interface{}); ok {
		for _, each := range untypedListens {
			if oneListen, ok := each.(string); ok {
				listens = append(listens, oneListen)
			} else {
				return nil, "", "", nil, NewErr{
					Code:    ErrConfigLoadType,
					value:   "listen",
					deepErr: fmt.Errorf("wrong type of %T - %#v", each, each),
				}
			}
		}
	} else {
		return nil, "", "", nil, NewErr{
			Code:    ErrConfigLoadStructure,
			value:   "listen",
			deepErr: fmt.Errorf("wrong type of %T - %#v", untypedListens, untypedListens),
		}
	}

	if accessLog, ok = c["accessLog"].(string); !ok {
		accessLog = ""
	}
	if errorLog, ok = c["errorLog"].(string); !ok {
		errorLog = ""
	}

	var handlers []HandlerConfig
	var dirtyHandlers []interface{}

	if dirtyHandlers, ok = c["handler"].([]interface{}); !ok {
		return nil, "", "", nil, NewErr{
			Code:    ErrConfigLoadStructure,
			value:   "handler",
			deepErr: fmt.Errorf("wrong type of %T", c["listen"]),
		}
	}

	for _, oneHandler := range dirtyHandlers {
		cleanHandler, err := decodeHandler(oneHandler)
		if err != nil {
			return nil, "", "", nil, err
		} else {
			handlers = append(handlers, cleanHandler)
		}
	}

	return listens, accessLog, errorLog, handlers, nil
}

func decodeHandler(dirtyHandler interface{}) (HandlerConfig, Err) {
	var h HandlerConfig

	var addressed map[string]interface{}
	var ok bool

	if addressed, ok = dirtyHandler.(map[string]interface{}); !ok {
		return HandlerConfig{}, NewErr{
			Code:    ErrConfigLoadStructure,
			value:   "handler",
			deepErr: fmt.Errorf("bad handler in config - %#v", dirtyHandler),
		}
	}

	if _, ok = addressed["handlerType"]; ok {
		if h.handlerType, ok = addressed["handlerType"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "handlerType",
			}
		}
	}
	if _, ok = addressed["handlerRoute"]; ok {
		if h.handlerRoute, ok = addressed["handlerRoute"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "handlerRoute",
			}
		}
	}

	if _, ok = addressed["baseURL"]; ok {
		if h.baseURL, ok = addressed["baseURL"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "baseURl",
			}
		}
	}
	if _, ok = addressed["confPath"]; ok {
		if h.confPath, ok = addressed["confPath"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "confPath",
			}
		}
	}
	if _, ok = addressed["confExt"]; ok {
		if h.confExt, ok = addressed["confExt"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "confExt",
			}
		}
	}

	if _, ok = addressed["exclude"]; ok {
		if h.exclude, ok = addressed["exclude"].([]string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "exclude",
			}
		}
	}

	if _, ok = addressed["subdomainLen"]; ok {
		if h.subdomainLen, ok = addressed["subdomainLen"].(int); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "subdomainLen",
			}
		}
	}

	if _, ok = addressed["redirectTracing"]; ok {
		if h.redirectTracing, ok = addressed["redirectTracing"].(bool); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadValue,
				value: "redirectTracing",
			}
		}
	}

	var err error
	if _, ok = addressed["confFile"]; ok {
		if workFile, ok := addressed["confFile"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:    ErrConfigLoadStructure,
				value:   "confFile",
				deepErr: fmt.Errorf("%#v", addressed["confFile"]),
			}
		} else if addressed["handlerType"] != "static" {
			h.confTempl, err = template.ParseFiles(workFile)
			if err != nil {
				return HandlerConfig{}, NewErr{
					Code:    ErrConfigLoadTemplate,
					value:   "confFile " + workFile,
					deepErr: err,
				}
			}
		}
	}
	if _, ok = addressed["resFile"]; ok {
		if workFile, ok := addressed["resFile"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadStructure,
				value: "resFile " + workFile,
			}
		} else if addressed["handlerType"] != "static" {
			h.resTempl, err = template.ParseFiles(workFile)
			if err != nil {
				return HandlerConfig{}, NewErr{
					Code:    ErrConfigLoadTemplate,
					value:   "resFile",
					deepErr: err,
				}
			}
		} else {
			h.resFile = workFile
		}
	}

	if _, ok = addressed["ipFile"]; ok {
		if workFile, ok := addressed["ipFile"].(string); !ok {
			return HandlerConfig{}, NewErr{
				Code:  ErrConfigLoadStructure,
				value: "ipFile " + workFile,
			}
		} else if addressed["ipFile"].(string) == "" {
		} else if addressed["handlerType"] != "static" {
			// #TODO# fix this call?
			h.ipList, err = parseIPList(workFile)
			if err != nil {
				return HandlerConfig{}, NewErr{
					Code:    ErrConfigLoadTemplate,
					value:   "ipFile",
					deepErr: err,
				}
			}
		}
	}

	return h, nil
}
