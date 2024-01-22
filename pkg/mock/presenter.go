package mock

type MockPresenter struct {
	WriteDataFn  func() error
	WriteErrorFn func(err error)
	WriteFn      func()
}

func (p *MockPresenter) WriteData() error {
	if p.WriteDataFn == nil {
		return nil
	}
	return p.WriteDataFn()
}

func (p *MockPresenter) WriteError(err error) {
	if p.WriteDataFn != nil {
		p.WriteErrorFn(err)
	}
}

func (p *MockPresenter) Write() {
	if p.WriteDataFn != nil {
		p.WriteFn()
	}
}
