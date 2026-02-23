package playwright

type fileChooserImpl struct {
	page          Page
	elementHandle ElementHandle
	isMultiple    bool
}

func (f *fileChooserImpl) Page() Page {
	return f.page
}

func (f *fileChooserImpl) Element() ElementHandle {
	return f.elementHandle
}

func (f *fileChooserImpl) IsMultiple() bool {
	return f.isMultiple
}

// InputFile represents the input file for:
// - FileChooser.SetFiles()
// - ElementHandle.SetInputFiles()
// - Page.SetInputFiles()
type InputFile struct {
	Name     string `json:"name"`
	MimeType string `json:"mimeType,omitempty"`
	Buffer   []byte `json:"buffer"`
}

func (f *fileChooserImpl) SetFiles(files any, options ...FileChooserSetFilesOptions) error {
	if len(options) == 1 {
		return f.elementHandle.SetInputFiles(files, ElementHandleSetInputFilesOptions(options[0]))
	}
	return f.elementHandle.SetInputFiles(files)
}

func newFileChooser(page Page, elementHandle ElementHandle, isMultiple bool) *fileChooserImpl {
	return &fileChooserImpl{
		page:          page,
		elementHandle: elementHandle,
		isMultiple:    isMultiple,
	}
}
