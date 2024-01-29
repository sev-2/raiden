package raiden

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

// ----- Presenter Contract ------
type Presenter interface {
	GetError() error
	GetData() any

	WriteData() error
	WriteError(err error)
	Write()
}

// ----- Define native presenter -----

// JSON Presenter
// handler http output response with json format
type JsonPresenter struct {
	response *fasthttp.Response
	data     any
	err      error
}

func NewJsonPresenter(response *fasthttp.Response) *JsonPresenter {
	return &JsonPresenter{
		response: response,
	}
}

func (p *JsonPresenter) SetData(data any) {
	p.data = data
}

func (p *JsonPresenter) SetError(err error) {
	if errResponse, ok := err.(*ErrorResponse); ok {
		p.response.SetStatusCode(errResponse.StatusCode)
	} else {
		p.response.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	p.err = err
}

func (p *JsonPresenter) GetError() error {
	return p.err
}

func (p *JsonPresenter) GetData() any {
	return p.data
}

func (p *JsonPresenter) WriteData() error {
	jStr, err := json.Marshal(p.data)
	if err != nil {
		return err
	}

	p.response.SetStatusCode(fasthttp.StatusOK)
	p.response.AppendBody(jStr)
	return nil
}

func (p JsonPresenter) WriteError(err error) {
	if errResponse, ok := err.(*ErrorResponse); ok {
		responseByte, errMarshall := json.Marshal(errResponse)
		if errMarshall == nil {
			p.response.SetStatusCode(errResponse.StatusCode)
			p.response.AppendBody(responseByte)
			return
		}
		err = errMarshall
	}
	p.response.SetStatusCode(fasthttp.StatusInternalServerError)
	p.response.AppendBodyString(err.Error())
}

func (p *JsonPresenter) Write() {
	p.response.Header.SetContentType("application/json")
	if p.err != nil {
		p.WriteError(p.err)
		return
	}

	if err := p.WriteData(); err != nil {
		p.err = err
		p.WriteError(err)
	}
}
