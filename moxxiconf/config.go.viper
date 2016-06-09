package moxxiConf

import (
	"fmt"
	"github.com/spf13/viper"
	"text/template"
)

func LoadConfig() ([]string, []HandlerConfig, Err) {
	config := prepConfig()

	if err := validateConfig(config); err != nil {
		return []string{}, []HandlerConfig{}, UpgradeError(err)
	}

	handlers, listens, err := loadConfig(config)
	if err != nil {
		return []string{}, []HandlerConfig{}, err
	}

	return handlers, listens, nil
}

func prepConfig() *viper.Viper {

	c := viper.New()
	// establish the config paths
	c.SetConfigName("config")
	c.AddConfigPath("/etc/moxxi/")
	c.AddConfigPath("$HOME/.moxxi")
	c.AddConfigPath(".")

	// set default values for the config
	c.SetDefault("listen", []string{":8080"})

	if err := config.ReadInConfig(); err != nil {
		return []string{}, []HandlerConfig{}, UpgradeError(err)
	}

	return c

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

func validateConfig(c *viper.Viper) Err {

	// clean up the listen line
	if _, ok := c.Get("listen").([]interface{}); !ok {
		c.Set("listen", []interface{}{c.Get("listen")})
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
	if c.IsSet("exclude") {
		if _, ok := c.Get("exclude").([]interface{}); !ok {
			c.Set("exclude", []interface{}{c.Get("exclude")})
			return NewErr{Code: ErrConfigBadStructure, value: "exclude"}
		}
	}
	if c.IsSet("subDomainLen") {
		if c.GetInt("subDomainLen") < 1 {
			return NewErr{Code: ErrConfigBadStructure, value: "subDomainLen"}
		}
	}

	// test and propagate handlers
	for i := 0; i < len(handlers); i++ {
		err := validateConfigHandler(c, fmt.Sprintf("handler.%d.", i))
		if err != nil {
			return err
		}
	}
	return nil
}

func validateConfigHandler(c *viper.Viper, base string) Err {
	// verify similar for individual handlers
	if !c.IsSet(base + "handlerType") {
		c.Set(base+"handlerType", "static")
	}

	if !c.IsSet(base + "handlerRoute") {
		return NewErr{
			Code:  ErrConfigBadStructure,
			value: base + "handlerRoute",
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
	if c.IsSet(base + "exclude") {
		if _, ok := c.Get(base + "exclude").([]string); !ok {
			return NewErr{Code: ErrConfigBadStructure, value: base + "exclude"}
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
		"exclude",
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
	return nil
}
