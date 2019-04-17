package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
)

type HandlerParams struct {
	Discriminator string      `json:"discriminator"`
	IsError       bool        `json:"is_error"`
	Data          interface{} `json:"data"`
}

type Handler struct {
	Method string          `json:"method"`
	Result []HandlerParams `json:"result"`
}

var configSchema string = `
{
  "definitions": {},
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/root.json",
  "type": "object",
  "title": "The Root Schema",
  "required": [
    "server_port",
    "url_prefix",
    "discriminator_header",
    "handlers"
  ],
  "properties": {
    "server_port": {
      "$id": "#/properties/server_port",
      "type": "integer",
      "title": "Server port",
      "default": 0,
      "examples": [
        8080
      ]
    },
    "url_prefix": {
      "$id": "#/properties/url_prefix",
      "type": "string",
      "title": "URL prefix",
      "default": "",
      "examples": [
        "v1"
      ],
      "pattern": "^(.*)$"
    },
    "discriminator_header": {
      "$id": "#/properties/discriminator_header",
      "type": "string",
      "title": "Discriminator HEADER",
      "default": "",
      "examples": [
        "JSON_RPC_MOCK"
      ],
      "pattern": "^(.*)$"
    },
    "handlers": {
      "$id": "#/properties/handlers",
      "type": "array",
      "title": "The Handlers Schema",
      "items": {
        "$id": "#/properties/handlers/items",
        "type": "object",
        "title": "The Items Schema",
        "required": [
          "method",
          "result"
        ],
        "properties": {
          "method": {
            "$id": "#/properties/handlers/items/properties/method",
            "type": "string",
            "title": "The Method Schema",
            "default": "",
            "examples": [
              "getAllProducts"
            ],
            "pattern": "^(.*)$"
          },
          "result": {
            "$id": "#/properties/handlers/items/properties/result",
            "type": "array",
            "title": "The Result Schema",
            "items": {
              "$id": "#/properties/handlers/items/properties/result/items",
              "type": "object",
              "title": "The Items Schema",
              "required": [
                "discriminator",
                "is_error",
                "data"
              ],
              "properties": {
                "discriminator": {
                  "$id": "#/properties/handlers/items/properties/result/items/properties/discriminator",
                  "type": "string",
                  "title": "The Discriminator Schema",
                  "default": "",
                  "examples": [
                    "success"
                  ],
                  "pattern": "^(.*)$"
                },
                "is_error": {
                  "$id": "#/properties/handlers/items/properties/result/items/properties/is_error",
                  "type": "boolean",
                  "title": "The Is_error Schema",
                  "default": false,
                  "examples": [
                    false
                  ]
                },
                "data": {
                  "$id": "#/properties/handlers/items/properties/result/items/properties/data",
                  "title": "The Data Schema",
                  "default": null
                }
              }
            }
          }
        }
      }
    }
  }
}
`

type Config struct {
	ServerPort          uint      `json:"server_port"`
	UrlPrefix           string    `json:"url_prefix"`
	DiscriminatorHeader string    `json:"discriminator_header"`
	Handlers            []Handler `json:"handlers"`
}

func (c *Config) Parse(configPath string) {
	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatal("Open config file error: ", err)
	}

	err = json.Unmarshal(configFile, c)
	if err != nil {
		logrus.Fatal("Config parse error: ", err)
	}
}

func (c *Config) Validate(configPath string) {
	configLoader := gojsonschema.NewReferenceLoader(fmt.Sprintf("file://%s", configPath))
	schemaLoader := gojsonschema.NewStringLoader(configSchema)

	validationResult, err := gojsonschema.Validate(schemaLoader, configLoader)
	if err != nil {
		logrus.Fatal("Config validation error: ", err)
	}

	if !validationResult.Valid() {
		fmt.Printf("Config is not valid. see errors :\n")
		for _, desc := range validationResult.Errors() {
			fmt.Printf("- %s\n", desc)
		}

		logrus.Fatal("Config validation error")
	}
}

func NewConfig() *Config {
	return &Config{}
}

type RequestBody struct {
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
	Id     int           `json:"id"`
}

type ResponseError struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Response struct {
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
	Id     int         `json:"id"`
}

func main() {
	configPath := flag.String("config", "", "Path to json rpc config")
	flag.Parse()

	if *configPath == "" {
		logrus.Fatal("Empty config path")
	}

	config := NewConfig()
	config.Validate(*configPath)
	config.Parse(*configPath)

	e := echo.New()

	e.Any(config.UrlPrefix, func(c echo.Context) error {
		body := &RequestBody{}
		err := c.Bind(body)
		if err != nil {
			return c.JSON(200, &Response{
				Error: &ResponseError{
					Code:  500,
					Error: "MOCK SERVER ERROR: parse body error",
				},
				Id: body.Id,
			})
		}

		discriminator := c.Request().Header.Get(config.DiscriminatorHeader)
		if discriminator == "" {
			discriminator = "success"
		}

		h := new(Handler)

		for _, handler := range config.Handlers {
			if handler.Method == body.Method {
				h = &handler

				break
			}
		}

		if h == nil {
			return c.JSON(200, &Response{
				Error: &ResponseError{
					Code:  500,
					Error: "MOCK SERVER ERROR: method not found",
				},
				Id: body.Id,
			})
		}

		p := new(HandlerParams)

		for _, params := range h.Result {
			if discriminator == params.Discriminator {
				p = &params

				break
			}
		}

		if p == nil {
			return c.JSON(200, &Response{
				Error: &ResponseError{
					Code:  500,
					Error: "MOCK SERVER ERROR: discriminator not found",
				},
				Id: body.Id,
			})
		}

		if p.IsError {
			return c.JSON(200, &Response{
				Error: p.Data,
				Id:    body.Id,
			})
		} else {
			return c.JSON(200, &Response{
				Result: p.Data,
				Id:     body.Id,
			})
		}
	})

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", config.ServerPort)))
}
