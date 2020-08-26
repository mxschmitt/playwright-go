package playwright

import (
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

var pw *Playwright

func TestMain(m *testing.M) {
	var err error
	pw, err = Run()
	if err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	code := m.Run()
	if err := pw.Stop(); err != nil {
		log.Fatalf("could not start Playwright: %v", err)
	}
	os.Exit(code)
}

type TestHelperData struct {
	Playwright *Playwright
	Browser    *Browser
	Context    *BrowserContext
	Page       *Page
	IsChromium bool
	IsFirefox  bool
	IsWebKit   bool
}

func (th *TestHelperData) Close(t *testing.T) {
	require.NoError(t, th.Browser.Close())
}

func NewTestHelper(t *testing.T) *TestHelperData {
	browserName := os.Getenv("BROWSER")
	var browserType *BrowserType
	if browserName == "chromium" || browserName == "" {
		browserType = pw.Chromium
	} else if browserName == "firefox" {
		browserType = pw.Firefox
	} else if browserName == "webkit" {
		browserType = pw.WebKit
	}

	browser, err := browserType.Launch(BrowserTypeLaunchOptions{
		Headless: Bool(os.Getenv("HEADFUL") == ""),
	})
	require.NoError(t, err)
	context, err := browser.NewContext()
	require.NoError(t, err)
	page, err := context.NewPage()
	require.NoError(t, err)
	return &TestHelperData{
		Playwright: pw,
		Browser:    browser,
		Context:    context,
		Page:       page,
		IsChromium: browserName == "chromium" || browserName == "",
		IsFirefox:  browserName == "firefox",
		IsWebKit:   browserName == "webkit",
	}
}

func Map(vs interface{}, f func(interface{}) interface{}) []interface{} {
	v := reflect.ValueOf(vs)
	vsm := make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		vsm[i] = f(v.Index(i).Interface())
	}
	return vsm
}

// ChanToSlice reads all data from ch (which must be a chan), returning a
// slice of the data. If ch is a 'T chan' then the return value is of type
// []T inside the returned interface.
// A typical call would be sl := ChanToSlice(ch).([]int)
func ChanToSlice(ch interface{}, amount int) interface{} {
	chv := reflect.ValueOf(ch)
	slv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(ch).Elem()), 0, 0)
	for i := 0; i < amount; i++ {
		v, ok := chv.Recv()
		if !ok {
			return slv.Interface()
		}
		slv = reflect.Append(slv, v)
	}
	return slv.Interface()
}
