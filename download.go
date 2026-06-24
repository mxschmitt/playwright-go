package playwright

type downloadImpl struct {
	page              Page
	url               string
	suggestedFilename string
	artifact          *artifactImpl
}

func (d *downloadImpl) String() string {
	return d.SuggestedFilename()
}

func (d *downloadImpl) Page() Page {
	return d.page
}

func (d *downloadImpl) URL() string {
	return d.url
}

func (d *downloadImpl) SuggestedFilename() string {
	return d.suggestedFilename
}

func (d *downloadImpl) Delete() error {
	err := d.artifact.Delete()
	return err
}

func (d *downloadImpl) Failure() error {
	return d.artifact.Failure()
}

func (d *downloadImpl) Path() (string, error) {
	path, err := d.artifact.PathAfterFinished()
	return path, err
}

func (d *downloadImpl) SaveAs(path string) error {
	err := d.artifact.SaveAs(path)
	return err
}

func (d *downloadImpl) Cancel() error {
	return d.artifact.Cancel()
}

func newDownload(page Page, url string, suggestedFilename string, artifact *artifactImpl) *downloadImpl {
	return &downloadImpl{
		page:              page,
		url:               url,
		suggestedFilename: suggestedFilename,
		artifact:          artifact,
	}
}
