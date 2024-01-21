package raiden

import (
	"encoding/json"
	"fmt"

	"github.com/valyala/fasthttp"
)

type Context struct {
	*fasthttp.RequestCtx
	config *Config
}

func (c *Context) GetConfig() *Config {
	return c.config
}

func (c *Context) SendJson(data any) {
	c.SetContentType("application/json")

	jStr, err := json.Marshal(data)
	if err != nil {
		c.SetStatusCode(fasthttp.StatusInternalServerError)
		c.WriteString(fmt.Sprintf("{\"message\" : %q}", err))
		return
	}

	c.SetStatusCode(fasthttp.StatusOK)
	c.Write(jStr)
}
