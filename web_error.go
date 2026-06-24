package playwright

type webErrorImpl struct {
	err      error
	page     Page
	location *WebErrorLocation
}

func (e *webErrorImpl) Page() Page {
	return e.page
}

func (e *webErrorImpl) Error() error {
	return e.err
}

func (e *webErrorImpl) Location() *WebErrorLocation {
	return e.location
}

func newWebError(page Page, err error, location *WebErrorLocation) WebError {
	return &webErrorImpl{
		err:      err,
		page:     page,
		location: location,
	}
}
