package raiden

import (
	"github.com/valyala/fasthttp"
)

type (
	Context interface {
		GetConfig() *Config
		SendData(data any) Presenter
		SendError(err error) Presenter
		SendErrorWithCode(statusCode int, err error) Presenter
	}

	context struct {
		*fasthttp.RequestCtx
		config *Config
	}
)

func (c *context) GetConfig() *Config {
	return c.config
}

func (c *context) SendData(data any) Presenter {
	presenter := NewJsonPresenter(c.RequestCtx)
	presenter.SetData(data)
	return presenter
}

func (c *context) SendError(err error) Presenter {
	presenter := NewJsonPresenter(c.RequestCtx)
	presenter.SetError(err)
	return presenter
}

func (c *context) SendErrorWithCode(statusCode int, err error) Presenter {
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
