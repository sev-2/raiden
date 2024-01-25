package raiden

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type Presenter interface {
	GetError() error

	WriteData() error
	WriteError(err error)
	Write()
}

type JsonPresenter struct {
	reqCtx *fasthttp.RequestCtx
	data   any
	err    error
}

func NewJsonPresenter(ctx *fasthttp.RequestCtx) *JsonPresenter {
	return &JsonPresenter{
		reqCtx: ctx,
	}
}

func (p *JsonPresenter) SetData(data any) {
	p.data = data
}

func (p *JsonPresenter) SetError(err error) {
	if errResponse, ok := err.(*ErrorResponse); ok {
		p.reqCtx.SetStatusCode(errResponse.StatusCode)
	} else {
		p.reqCtx.SetStatusCode(fasthttp.StatusInternalServerError)
	}
	p.err = err
}

func (p *JsonPresenter) GetError() error {
	return p.err
}

func (p *JsonPresenter) WriteData() error {
	jStr, err := json.Marshal(p.data)
	if err != nil {
		return err
	}

	p.reqCtx.SetStatusCode(fasthttp.StatusOK)
	p.reqCtx.Write(jStr)
	return nil
}

func (p JsonPresenter) WriteError(err error) {
	if errResponse, ok := err.(*ErrorResponse); ok {
		responseByte, errMarshall := json.Marshal(errResponse)
		if errMarshall == nil {
			p.reqCtx.SetStatusCode(errResponse.StatusCode)
			p.reqCtx.Write(responseByte)
			return
		}
		err = errMarshall
	}
	p.reqCtx.SetStatusCode(fasthttp.StatusInternalServerError)
	p.reqCtx.WriteString(err.Error())
}

func (p *JsonPresenter) Write() {
	p.reqCtx.SetContentType("application/json")
	if p.err != nil {
		p.WriteError(p.err)
		return
	}

	if err := p.WriteData(); err != nil {
		p.err = err
		p.WriteError(err)
	}
}
