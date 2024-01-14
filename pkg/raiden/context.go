package raiden

import "github.com/labstack/echo/v4"

type Context struct {
	echo.Context
	config *Config
}

func (c *Context) GetConfig() *Config {
	return c.config
}
