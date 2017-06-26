package moxxiConf

import (
	"bytes"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func ExamplePrepConfig() {
	c := prepConfig()
	c.Debug()
	// Output:
	// 	Aliases:
	// Aliases:
	// map[string]string{}
	// Override:
	// map[string]interface {}{}
	// PFlags:
	// map[string]viper.FlagValue{}
	// Env:
	// map[string]string{}
	// Key/Value Store:
	// map[string]interface {}{}
	// Config:
	// map[string]interface {}{}
	// Defaults:
	// map[string]interface {}{"listen":[]string{":8080"}}
}

func TestValidateConfig(t *testing.T) {
	t.Skip("skipped config validation test")
	c := viper.New()
	c.SetConfigType("json")

	// any approach to require this configuration into your program.
	var config = []byte(`
{
  "listen": [
    "localhost:8080",
    "localhost:8081"
  ],
  "baseURL": "parentdomain.com",
  "confPath": "/home/moxxi/vhosts.d",
  "confExt": ".conf",
  "confFile": "/home/moxxi/vhost.template",
  "resFile": "/home/moxxi/response.template",
  "subdomainLen": 8,
  "exclude": [
    "parentdomain.com",
    "moxxi.parentdomain.com",
    "notallowed.parentdomain.com"
  ],
  "handler": [
    {
      "handlerType": "static",
      "handlerRoute": "/",
      "resFile": "/home/moxxi/static_form.html"
    },
    {
      "handlerType": "form",
      "handlerRoute": "/submit"
    },
    {
      "handlerType": "json",
      "handlerRoute": "/json"
    }
  ]
}
	`)

	globErr := c.ReadConfig(bytes.NewBuffer(config))
	if globErr != nil {
		log.Printf("got an error I should not have\n%s\ndumping config",
			globErr.Error())
		c.Debug()
		t.FailNow()
	}

	intErr := validateConfig(c)
	if intErr != nil {
		log.Printf("got an error I should not have\n%s\ndumping config",
			intErr.Error())
		c.Debug()
		t.FailNow()
	}

	var dataInspection = []struct {
		location string
		expected interface{}
	}{
		{
			location: "listen",
			expected: []string{
				"localhost:8080",
				"localhost:8081",
			},
		}, {
			location: "handler.1.confPath",
			expected: "/home/moxxi/vhosts.d",
		}, {
			location: "handler.2.confPath",
			expected: "/home/moxxi/vhosts.d",
		}, {
			location: "handler.0.resFile",
			expected: "/home/moxxi/static_form.html",
		}, {
			location: "handler.0.resFile",
			expected: "/home/moxxi/response.template",
		}, {
			location: "handler.1.subdomainLen",
			expected: 8,
		}, {
			location: "handler.2.exclude",
			expected: []string{
				"parentdomain.com",
				"moxxi.parentdomain.com",
				"notallowed.parentdomain.com",
			},
		},
	}

	for _, test := range dataInspection {
		assert.Equal(t, test.expected, c.Get(test.location), "got the wrong value at location %s", test.location)
		// switch test.expected.(type) {
		// case string:
		// case []string:
		// case int:

		// }
	}
}
