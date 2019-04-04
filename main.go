package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"
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

type Config struct {
	ServerPort uint      `json:"server_port"`
	UrlPrefix  string    `json:"url_prefix"`
	Handlers   []Handler `json:"handlers"`
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

		discriminator := c.Request().Header.Get("JSON_RPC_MOCK")
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
