package playwright_test

import (
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestConsoleShouldWork(t *testing.T) {
	BeforeEach(t)

	messages := make(chan playwright.ConsoleMessage, 1)
	page.Once("console", func(message playwright.ConsoleMessage) {
		messages <- message
	})
	_, err := page.Evaluate(`() => console.log("hello", 5, {foo: "bar"})`)
	require.NoError(t, err)
	message := <-messages
	if !isFirefox {
		require.Equal(t, message.Text(), "hello 5 {foo: bar}")
		require.Equal(t, message.String(), "hello 5 {foo: bar}")
	} else {
		require.Equal(t, message.Text(), "hello 5 JSHandle@object")
		require.Equal(t, message.String(), "hello 5 JSHandle@object")
	}
	require.Equal(t, message.Type(), "log")
	jsonValue1, err := message.Args()[0].JSONValue()
	require.NoError(t, err)
	require.Equal(t, "hello", jsonValue1)
	jsonValue2, err := message.Args()[1].JSONValue()
	require.NoError(t, err)
	require.Equal(t, 5, jsonValue2)
	jsonValue3, err := message.Args()[2].JSONValue()
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{
		"foo": "bar",
	}, jsonValue3)
}

func TestConsoleShouldEmitSameLogTwice(t *testing.T) {
	BeforeEach(t)

	messages := make(chan string, 2)
	page.OnConsole(func(message playwright.ConsoleMessage) {
		messages <- message.Text()
	})
	_, err := page.Evaluate(`() => { for (let i = 0; i < 2; ++i ) console.log("hello"); } `)
	require.NoError(t, err)
	m1 := <-messages
	m2 := <-messages
	require.Equal(t, []string{"hello", "hello"}, []string{m1, m2})
}

func TestConsoleShouldUseTextForStr(t *testing.T) {
	BeforeEach(t)

	messages := make(chan playwright.ConsoleMessage, 1)
	page.OnConsole(func(message playwright.ConsoleMessage) {
		messages <- message
	})
	_, err := page.Evaluate(`() => console.log("Hello world")`)
	require.NoError(t, err)
	message := <-messages
	require.Equal(t, "Hello world", message.String())
}

func TestConsoleShouldWorkForDifferentConsoleAPICalls(t *testing.T) {
	BeforeEach(t)

	messagesChan := make(chan playwright.ConsoleMessage, 6)
	page.OnConsole(func(message playwright.ConsoleMessage) {
		messagesChan <- message
	})
	// All console events will be reported before 'page.evaluate' is finished.
	_, err := page.Evaluate(
		`() => {
      // A pair of time/timeEnd generates only one Console API call.
      console.time('calling console.time');
      console.timeEnd('calling console.time');
      console.trace('calling console.trace');
      console.dir('calling console.dir');
      console.warn('calling console.warn');
      console.error('calling console.error');
      console.log(Promise.resolve('should not wait until resolved!'));
	}`,
	)
	messages := ChanToSlice(messagesChan, 6)
	require.NoError(t, err)
	require.Equal(t, []string{
		"timeEnd",
		"trace",
		"dir",
		"warning",
		"error",
		"log",
	}, Map(messages, func(msg playwright.ConsoleMessage) string {
		return msg.Type()
	}))

	require.Contains(t, messages[0].Text(), "calling console.time")
	require.Equal(t, []string{
		"calling console.trace",
		"calling console.dir",
		"calling console.warn",
		"calling console.error",
		"Promise",
	}, Map(messages[1:], func(msg playwright.ConsoleMessage) string {
		return msg.Text()
	}))
}

func TestConsoleShouldNotFailForWindowObjects(t *testing.T) {
	BeforeEach(t)

	messages := make(chan playwright.ConsoleMessage, 1)
	page.Once("console", func(message playwright.ConsoleMessage) {
		messages <- message
	})
	_, err := page.Evaluate("() => console.error(window)")
	require.NoError(t, err)
	message := <-messages
	if !isFirefox {
		require.Equal(t, "Window", message.Text())
	} else {
		require.Equal(t, "JSHandle@object", message.Text())
	}
}

func TestConsoleShouldTriggerCorrectLog(t *testing.T) {
	BeforeEach(t)

	messages := make(chan playwright.ConsoleMessage, 1)
	page.Once("console", func(message playwright.ConsoleMessage) {
		messages <- message
	})
	_, err := page.Goto("about:blank")
	require.NoError(t, err)
	_, err = page.Evaluate("url => fetch(url).catch(e => {})", server.EMPTY_PAGE)
	require.NoError(t, err)
	message := <-messages

	headerString := "Access-Control-Allow-Origin"
	corsString := "CORS"
	require.Condition(t, func() bool {
		return strings.Contains(message.Text(), headerString) || strings.Contains(message.Text(), corsString)
	}, "The text should contain either '%s' or '%s'", headerString, corsString)

	require.Equal(t, "error", message.Type())
}

func TestConsoleShouldHaveLocationForConsoleAPICalls(t *testing.T) {
	BeforeEach(t)

	messageEvent, err := page.ExpectEvent("console", func() error {
		_, err := page.Goto(server.PREFIX + "/consolelog.html")
		return err
	}, playwright.PageExpectEventOptions{
		Predicate: func(m playwright.ConsoleMessage) bool {
			return strings.HasPrefix(m.Text(), "here:")
		},
	})
	require.NoError(t, err)
	message := messageEvent.(playwright.ConsoleMessage)
	require.Equal(t, message.Type(), "log")
	require.Equal(t, server.PREFIX+"/consolelog.html", message.Location().URL)
	require.Equal(t, 7, message.Location().LineNumber)
	// The non-deprecated Line/Column aliases must be populated too (upstream
	// maps lineNumber/columnNumber onto line/column).
	require.Equal(t, message.Location().LineNumber, message.Location().Line)
	require.Equal(t, message.Location().ColumnNumber, message.Location().Column)
}
