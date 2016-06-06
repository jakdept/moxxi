package MoxxiConf


func prepConfig() *viper.Viper {

	c := viper.New()
	// establish the config paths
	c.SetConfigName("config")
	c.AddConfigPath("/etc/moxxi/")
	c.AddConfigPath("$HOME/.moxxi")
	c.AddConfigPath(".")

	// set default values for the config
	c.SetDefault("listen", []string{":8080"})

	return &c

}

func loadConfig(c *viper.Viper) ([]string, []HandlerConfig, Err) {

	viper.Debug()

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

	return handlers, listens, nil
}

func verifyConfig(c *viper.Viper) Err {

	// clean up the listen line
	switch c.Get("listen").(type) {
	case []string:
	case string:
		c.Set("listen", []string{c.GetString("listen")})
	default:
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "listen",
		}
	}

	// verify the handlers are an array
	handlers, ok := c.Get("handler").([]interface{})
	if !ok {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: "handler",
		}
	}

	// check that the following things are strings in the top level
	for _, subConf := range []string{
		"baseURL",
		"confFile",
		"confExt",
		"confFile",
		"resFile",
	} {
		if c.IsSet(subConf) {
			if _, ok := c.Get(subConf).(string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: subConf}
			}
		}
	}

	// toplevel other typechecks
	if c.IsSet("excludes") {
		if _, ok := c.Get("excludes").([]string); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: "excludes"}
		}
	}
	if c.IsSet("subDomainLen") {
		if _, ok := c.Get("subDomainLen").(int); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: "subDomainLen"}
		}
	}

	// verify similar for individual handlers
	for i := 0; i < len(handlers); i++ {
		base := fmt.Sprintf("handler.%d.", i)

		if !c.IsSet(base + "handlerType") {
			c.Set(base+".handlerType", "static")
		}

		if !c.IsSet(base + "handlerRoute") {
			return NewErr{
				Code:  ErrConfigBadValue,
				value: "handlerRoute",
			}
		}

		// string typecheck
		for _, subConf := range []string{
			"baseURL",
			"confFile",
			"confExt",
			"confFile",
			"resFile",
		} {
			if c.IsSet(base + subConf) {
				if _, ok := c.Get(base + subConf).(string); !ok {
					return NewErr{Code: ErrConfigBadStructure, value: base + subConf}
				}
			}
		}

		// other typechecks
		if c.IsSet(base + "excludes") {
			if _, ok := c.Get(base + "excludes").([]string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: base + "excludes"}
			}
		}
		if c.IsSet(base + "subdomainLen") {
			if _, ok := c.Get(base + "subdomainLen").([]string); !ok {
				return NewErr{Code: ErrConfigBadStructure, value: base + "subdomainLen"}
			}
		}

		// propagate unset values from the global setting
		for _, subConf := range []string{
			"baseURL",
			"confFile",
			"confExt",
			"confFile",
			"resFile",
			"excludes",
			"subdomainLen",
		} {
			if !c.IsSet(base+subConf) && c.IsSet(subConf) {
				c.Set(base+subConf, c.Get(subConf))
			}
		}

		// parse the two templates
		if c.GetString(base+"handlerType") != "static" && c.IsSet(base+"confFile") {
			template, err := template.ParseFiles(c.GetString(base + "confFile"))
			if err != nil {
				return NewErr{
					Code:    ErrConfigBadTemplate,
					value:   c.GetString(base + "handlerType"),
					deepErr: err,
				}
			}
			c.Set(base+"confTempl", template)
		}

		if c.GetString(base+"handlerType") != "static" && c.IsSet(base+"resFile") {
			template, err := template.ParseFiles(c.GetString(base + "resFile"))
			if err != nil {
				return NewErr{
					Code:    ErrConfigBadTemplate,
					value:   c.GetString(base + "handlerType"),
					deepErr: err,
				}
			}
			c.Set(base+"resTempl", template)
		}
	}
}

func LoadConfig() ([]string, []HandlerConfig, Err) {
	config := prepConfig()

	// #TODO# clean this
	err := config.ReadInConfig()
	if err != nil {
		log.Printf("Fatal error config file: %s \n", err)
		return MoxxiConf{}, err
	}

	if err = verifyConfig(config); err != nil {
		return []string{}, []HandlerConfig{}, UpgradeError(err)
	}

	listens, handlers, err := loadConfig()
	if err != nil {
		return listens, handlers, err
	}

	return config, nil
}