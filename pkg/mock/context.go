package mock

import "github.com/sev-2/raiden"

type MockContext struct {
	GetConfigFn         func() *raiden.Config
	SendDataFn          func(data any) raiden.Presenter
	SendErrorFn         func(err error) raiden.Presenter
	SendErrorWithCodeFn func(statusCode int, err error) raiden.Presenter
}

func (c *MockContext) GetConfig() *raiden.Config {
	return c.GetConfigFn()
}

func (c *MockContext) SendData(data any) raiden.Presenter {
	return c.SendDataFn(data)
}

func (c *MockContext) SendError(err error) raiden.Presenter {
	return c.SendErrorFn(err)
}

func (c *MockContext) SendErrorWithCode(statusCode int, err error) raiden.Presenter {
	return c.SendErrorWithCodeFn(statusCode, err)
}
