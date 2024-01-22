package raiden

import (
	"github.com/valyala/fasthttp"
)

type Context struct {
	*fasthttp.RequestCtx
	config *Config
}

func (c *Context) GetConfig() *Config {
	return c.config
}

func (c *Context) SendData(data any) *JsonPresenter {
	presenter := NewJsonPresenter(c.RequestCtx)
	presenter.SetData(data)
	return presenter
}

func (c *Context) SendError(err error) *JsonPresenter {
	presenter := NewJsonPresenter(c.RequestCtx)
	presenter.SetError(err)
	return presenter
}

func (c *Context) SendErrorWithCode(statusCode int, err error) *JsonPresenter {
	presenter := NewJsonPresenter(c.RequestCtx)
	if errResponse, ok := err.(*ErrorResponse); ok {
		errResponse.StatusCode = statusCode
		err = errResponse
	} else {
		errResponse := &ErrorResponse{
			Message:    err.Error(),
			StatusCode: statusCode,
			Code:       fasthttp.StatusMessage(statusCode),
		}
		err = errResponse
	}

	presenter.SetError(err)
	return presenter
}
