package playwright_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestBrowserContextExposeBinding(t *testing.T) {
	BeforeEach(t)

	bindingSource := []*playwright.BindingSource{}
	binding := func(source *playwright.BindingSource, a, b int) int {
		bindingSource = append(bindingSource, source)
		return a + b
	}
	err := context.ExposeBinding("add", func(source *playwright.BindingSource, args ...interface{}) interface{} {
		return binding(source, args[0].(int), args[1].(int))
	})
	require.NoError(t, err)
	page, err := context.NewPage()
	require.NoError(t, err)
	result, err := page.Evaluate("add(5, 6)")
	require.NoError(t, err)
	require.Equal(t, bindingSource[0].Context, context)
	require.Equal(t, bindingSource[0].Page, page)
	require.Equal(t, bindingSource[0].Frame, page.MainFrame())
	require.Equal(t, result, 11)
}

func TestBrowserContextExposeFunction(t *testing.T) {
	BeforeEach(t)

	err := context.ExposeFunction("add", func(args ...interface{}) interface{} {
		return args[0].(int) + args[1].(int)
	})
	require.NoError(t, err)
	page, err := context.NewPage()
	require.NoError(t, err)
	err = page.ExposeFunction("mul", func(args ...interface{}) interface{} {
		return args[0].(int) * args[1].(int)
	})
	require.NoError(t, err)
	err = context.ExposeFunction("sub", func(args ...interface{}) interface{} {
		return args[0].(int) - args[1].(int)
	})
	require.NoError(t, err)
	result, err := page.Evaluate(`async function() {
		return { mul: await mul(9, 4), add: await add(9, 4), sub: await sub(9, 4) }
	  }`)
	require.NoError(t, err)
	resultMap := result.(map[string]interface{})
	require.Equal(t, 36, resultMap["mul"])
	require.Equal(t, 13, resultMap["add"])
	require.Equal(t, 5, resultMap["sub"])
}

func TestBrowserContextExposeBindingPanic(t *testing.T) {
	BeforeEach(t)

	err := context.ExposeBinding("woof", func(source *playwright.BindingSource, args ...interface{}) interface{} {
		panic(errors.New("WOOF WOOF"))
	})
	require.NoError(t, err)
	page, err := context.NewPage()
	require.NoError(t, err)
	result, err := page.Evaluate(`async () => {
		try {
		  await window['woof']();
		} catch (e) {
		  return {message: e.message, stack: e.stack};
		}
	  }`)
	require.NoError(t, err)
	innerError := result.(map[string]interface{})
	require.Equal(t, innerError["message"], "WOOF WOOF")
	stack := strings.Split(innerError["stack"].(string), "\n")
	require.Contains(t, stack[4], "binding_test.go")
}

func TestBrowserContextExposeBindingObjectShouldBePassedByValue(t *testing.T) {
	BeforeEach(t)

	targets := []map[string]interface{}{}

	logme := func(t interface{}) int {
		targets = append(targets, t.(map[string]interface{}))
		return 17
	}

	err := page.ExposeBinding("logme", func(source *playwright.BindingSource, args ...interface{}) interface{} {
		return logme(args[0])
	})
	require.NoError(t, err)
	result, err := page.Evaluate("logme({ foo: 42 })")
	require.NoError(t, err)
	require.Equal(t, result, 17)
	require.Equal(t, 42, targets[0]["foo"])
}

func TestPageExposeBindingPanic(t *testing.T) {
	BeforeEach(t)

	err := page.ExposeBinding("woof", func(source *playwright.BindingSource, args ...interface{}) interface{} {
		panic(errors.New("WOOF WOOF"))
	})
	require.NoError(t, err)
	result, err := page.Evaluate(`async () => {
		try {
		  await window['woof']();
		} catch (e) {
		  return {message: e.message, stack: e.stack};
		}
	  }`)
	require.NoError(t, err)
	innerError := result.(map[string]interface{})
	require.Equal(t, innerError["message"], "WOOF WOOF")
	stack := strings.Split(innerError["stack"].(string), "\n")
	require.Contains(t, stack[4], "binding_test.go")
}

// TestPageExposeBindingPanicNonError verifies a binding that panics with a
// non-error value (e.g. a string) rejects cleanly instead of crashing the
// process on a failed type assertion in the recover handler.
func TestPageExposeBindingPanicNonError(t *testing.T) {
	BeforeEach(t)

	err := page.ExposeBinding("woofString", func(source *playwright.BindingSource, args ...interface{}) interface{} {
		panic("WOOF STRING")
	})
	require.NoError(t, err)
	result, err := page.Evaluate(`async () => {
		try {
		  await window['woofString']();
		} catch (e) {
		  return {message: e.message};
		}
	  }`)
	require.NoError(t, err)
	innerError := result.(map[string]interface{})
	require.Equal(t, "WOOF STRING", innerError["message"])
}

func TestPageBindingsNoRace(t *testing.T) {
	BeforeEach(t)

	wg := &sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		i := i
		go func() {
			defer wg.Done()
			err := page.ExposeBinding(fmt.Sprintf("foo%d", i), func(source *playwright.BindingSource, args ...interface{}) interface{} {
				return 42
			})
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	ret, err := page.Evaluate(`async () => {
		try {
		  return await window['foo9']();
		} catch (e) {
		  return {message: e.message, stack: e.stack};
		}
	  }`)
	require.NoError(t, err)
	require.Equal(t, 42, ret)
}
