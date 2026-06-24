package playwright_test

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestJSHandleGetProperty(t *testing.T) {
	BeforeEach(t)

	aHandle, err := page.EvaluateHandle(`() => ({
		one: 1,
		two: 2,
		three: 3
	})`)
	require.NoError(t, err)
	twoHandle, err := aHandle.GetProperty("two")
	require.NoError(t, err)
	value, err := twoHandle.JSONValue()
	require.NoError(t, err)
	require.Equal(t, value, 2)
}

func TestJSHandleGetProperties(t *testing.T) {
	BeforeEach(t)

	aHandle, err := page.EvaluateHandle(`() => ({
		one: 1,
		two: 2,
		three: 3
	})`)
	require.NoError(t, err)
	twoHandle, err := aHandle.GetProperties()
	require.NoError(t, err)
	v1, err := twoHandle["one"].JSONValue()
	require.NoError(t, err)
	require.Equal(t, 1, v1)

	v1, err = twoHandle["two"].JSONValue()
	require.NoError(t, err)
	require.Equal(t, 2, v1)

	v1, err = twoHandle["three"].JSONValue()
	require.NoError(t, err)
	require.Equal(t, 3, v1)
}

func TestJSHandleEvaluate(t *testing.T) {
	BeforeEach(t)

	aHandle, err := page.EvaluateHandle(`() => ({
		one: 1,
		two: 2,
		three: 3
	})`)
	require.NoError(t, err)
	twoHandle, err := aHandle.Evaluate("x => x")
	require.NoError(t, err)
	values := twoHandle.(map[string]interface{})
	require.Equal(t, 1, values["one"])
	require.Equal(t, 2, values["two"])
	require.Equal(t, 3, values["three"])
}

func TestJSHandleEvaluateHandle(t *testing.T) {
	BeforeEach(t)

	aHandle, err := page.EvaluateHandle(`() => ({
		one: 1,
		two: 2,
		three: 3
	})`)
	require.NoError(t, err)
	twoHandle, err := aHandle.EvaluateHandle("x => x")
	require.NoError(t, err)
	twoHandle, err = twoHandle.GetProperty("two")
	require.NoError(t, err)
	value, err := twoHandle.JSONValue()
	require.NoError(t, err)
	require.Equal(t, value, 2)
}

func TestJSHandleEvaluateHandleReturningElement(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	inner, err := page.QuerySelector("#inner") // nolint: staticcheck
	require.NoError(t, err)

	// EvaluateHandle may return an ElementHandle (e.g. element.parentElement).
	// It must not panic when the underlying handle is an ElementHandle.
	// See https://github.com/playwright-community/playwright-go/issues/332
	parent, err := inner.EvaluateHandle("element => element.parentElement")
	require.NoError(t, err)
	require.NotNil(t, parent.AsElement())

	id, err := parent.AsElement().GetAttribute("id") // nolint: staticcheck
	require.NoError(t, err)
	require.Equal(t, "outer", id)

	// GetProperty must likewise survive returning an ElementHandle.
	childHandle, err := parent.GetProperty("firstElementChild")
	require.NoError(t, err)
	require.NotNil(t, childHandle.AsElement())
}

func TestJSHandleGetPropertiesReturningElement(t *testing.T) {
	BeforeEach(t)

	_, err := page.Goto(server.PREFIX + "/dom.html")
	require.NoError(t, err)
	inner, err := page.QuerySelector("#inner") // nolint: staticcheck
	require.NoError(t, err)

	parent, err := inner.EvaluateHandle("element => ({ parent: element.parentElement })")
	require.NoError(t, err)
	props, err := parent.GetProperties()
	require.NoError(t, err)
	require.NotNil(t, props["parent"].AsElement())
}

func TestJSHandleTypeParsing(t *testing.T) {
	BeforeEach(t)

	aHandle, err := page.EvaluateHandle(`() => ({
		an_integer: 1,
		a_float: 2.2222222222,
		a_string_of_an_integer: "3",
	})`)
	require.NoError(t, err)
	twoHandle, err := aHandle.EvaluateHandle("x => x")
	require.NoError(t, err)

	integerHandle, err := twoHandle.GetProperty("an_integer")
	require.NoError(t, err)
	intV, err := integerHandle.JSONValue()
	require.NoError(t, err)
	_, ok := intV.(int)
	require.True(t, ok)
	_, ok = intV.(float64)
	require.False(t, ok)

	floatHandle, err := twoHandle.GetProperty("a_float")
	require.NoError(t, err)
	floatV, err := floatHandle.JSONValue()
	require.NoError(t, err)
	_, ok = floatV.(float64)
	require.True(t, ok)
	_, ok = floatV.(int)
	require.False(t, ok)

	stringHandle, err := twoHandle.GetProperty("a_string_of_an_integer")
	require.NoError(t, err)
	stringV, err := stringHandle.JSONValue()
	require.NoError(t, err)
	_, ok = stringV.(string)
	require.True(t, ok)
	_, ok = stringV.(int)
	require.False(t, ok)
}

func TestEvaluate(t *testing.T) {
	BeforeEach(t)

	t.Run("should work", func(t *testing.T) {
		val, err := page.Evaluate(`7 * 3`)
		require.NoError(t, err)
		require.Equal(t, 21, val)
	})

	t.Run("return nil for null", func(t *testing.T) {
		val, err := page.Evaluate(`() => null`)
		require.NoError(t, err)
		require.Nil(t, val)
	})

	t.Run("return nil for nil", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, nil)
		require.NoError(t, err)
		require.Nil(t, val)
	})

	t.Run("transfer nan", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, math.NaN())
		require.NoError(t, err)
		require.True(t, math.IsNaN(val.(float64)))
	})

	t.Run("transfer neg zero", func(t *testing.T) {
		// https://github.com/golang/go/issues/2196
		nz := math.Copysign(0, -1)
		val, err := page.Evaluate(`a => a`, nz)
		require.NoError(t, err)
		require.Equal(t, nz, val)
	})

	t.Run("transfer infinity", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, math.Inf(1))
		require.NoError(t, err)
		require.Equal(t, math.Inf(1), val)
	})

	t.Run("transfer -infinity", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, math.Inf(-1))
		require.NoError(t, err)
		require.Equal(t, math.Inf(-1), val)
	})

	t.Run("transfer roundtrip unserializable values", func(t *testing.T) {
		value := map[string]interface{}{"inf": math.Inf(1), "ninf": math.Inf(-1), "nzero": math.Copysign(0, -1)}
		val, err := page.Evaluate(`a => a`, value)
		require.NoError(t, err)
		require.Equal(t, value, val)
	})

	t.Run("transfer slices", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, []string{"test1", "test2", "test3"})
		require.NoError(t, err)
		require.Equal(t, []interface{}{"test1", "test2", "test3"}, val)

		val, err = page.Evaluate(`a => a`, []int{1, 2, 3})
		require.NoError(t, err)
		require.Equal(t, []interface{}{1, 2, 3}, val)
	})

	t.Run("transfer bigint", func(t *testing.T) {
		val, err := page.Evaluate(`() => 42n`)
		require.NoError(t, err)
		require.Equal(t, val, big.NewInt(42))
		val, err = page.Evaluate(`a => a`, big.NewInt(17))
		require.NoError(t, err)
		require.Equal(t, val, big.NewInt(17))
	})

	t.Run("return undefined for objects with symbols", func(t *testing.T) {
		val, err := page.Evaluate(`[Symbol("foo4")]`)
		require.NoError(t, err)
		require.Equal(t, []interface{}{nil}, val)

		val, err = page.Evaluate(`() => {
				const a = { };
				a[Symbol('foo4')] = 42;
				return a;
			}`)
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{}, val)

		val, err = page.Evaluate(`() => {
			return { foo: [{ a: Symbol('foo4') }] };
		}`)
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{"foo": []interface{}{map[string]interface{}{"a": nil}}}, val)
	})

	t.Run("work with unicode chars", func(t *testing.T) {
		val, err := page.Evaluate(`a => a["中文字符"]`, map[string]interface{}{"中文字符": 42})
		require.NoError(t, err)
		require.Equal(t, 42, val)
	})

	t.Run("throw when evaluation triggers reload", func(t *testing.T) {
		_, err := page.Evaluate(`() => { location.reload(); return new Promise(() => {}); }`)
		require.ErrorContains(t, err, "navigation")
	})

	t.Run("work with exposed function", func(t *testing.T) {
		require.NoError(t, page.ExposeFunction("callController", func(args ...interface{}) interface{} {
			return args[0].(int) * args[1].(int)
		}))
		val, err := page.Evaluate(`callController(9, 3)`)
		require.NoError(t, err)
		require.Equal(t, 27, val)
	})

	t.Run("reject promise with exception", func(t *testing.T) {
		_, err := page.Evaluate(`not_existing_object.property`)
		require.ErrorContains(t, err, "not_existing_object")
	})

	t.Run("support thrown strings", func(t *testing.T) {
		_, err := page.Evaluate(`() => { throw "qwerty" }`)
		require.ErrorContains(t, err, "qwerty")
	})

	t.Run("support thrown numbers", func(t *testing.T) {
		_, err := page.Evaluate(`() => { throw "100500" }`)
		require.ErrorContains(t, err, "100500")
	})

	t.Run("return complex objects", func(t *testing.T) {
		obj := map[string]interface{}{
			"foo": "bar!",
		}
		val, err := page.Evaluate(`a => a`, obj)
		require.NoError(t, err)
		require.Equal(t, obj, val)
	})

	t.Run("accept nil as one of multiple parameters", func(t *testing.T) {
		val, err := page.Evaluate(`({ a, b }) => Object.is(a, null) && Object.is(b, "foo")`, map[string]interface{}{"a": nil, "b": "foo"})
		require.NoError(t, err)
		require.True(t, val.(bool))
	})

	t.Run("properly serialize nil arguments", func(t *testing.T) {
		val, err := page.Evaluate(`x => ({a: x})`, nil)
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{"a": nil}, val)
	})

	t.Run("alias window document and node", func(t *testing.T) {
		val, err := page.Evaluate(`[window, document, document.body]`)
		require.NoError(t, err)
		require.Equal(t, []interface{}{"ref: <Window>", "ref: <Document>", "ref: <Node>"}, val)
	})

	t.Run("should work for circular object", func(t *testing.T) {
		val, err := page.Evaluate(`() => {
			const a = {x: 47};
			const b = {a};
			a.b = b;
			return a;
		}`)
		require.NoError(t, err)
		result := val.(map[string]interface{})
		require.Equal(t, 47, result["x"])
		b := result["b"].(map[string]interface{})
		a := b["a"]
		require.Equal(t, 47, a.(map[string]interface{})["x"])
	})

	t.Run("accept string", func(t *testing.T) {
		val, err := page.Evaluate("1 + 2")
		require.NoError(t, err)
		require.Equal(t, 3, val)
	})

	t.Run("throw if underlying element was disposed", func(t *testing.T) {
		require.NoError(t, page.SetContent("<section>39</section>"))
		element, err := page.QuerySelector("section") // nolint: staticcheck
		require.NoError(t, err)
		require.NoError(t, element.Dispose())
		_, err = page.Evaluate("e => e.textContent", element)
		require.ErrorContains(t, err, "no object with guid")
	})

	t.Run("evaluate exception", func(t *testing.T) {
		val, err := page.Evaluate(`() => {
			function innerFunction() {
				const e = new Error('error message');
				e.name = 'foobar';
				return e;
			}
			return innerFunction();
		}`)
		require.NoError(t, err)
		var e *playwright.Error
		require.True(t, errors.As(val.(error), &e))
		require.Equal(t, "foobar", e.Name)
		require.Equal(t, "error message", e.Message)
		require.Contains(t, e.Stack, "innerFunction")
	})

	t.Run("pass exception argument", func(t *testing.T) {
		eee := &playwright.Error{
			Name:    "foobar",
			Message: "error message",
			Stack:   "test stack",
		}
		val, err := page.Evaluate(`e => {
			return { message: e.message, name: e.name, stack: e.stack };
		}`, eee)
		require.NoError(t, err)
		require.Equal(t, "foobar", val.(map[string]interface{})["name"])
		require.Equal(t, "error message", val.(map[string]interface{})["message"])
		require.Contains(t, val.(map[string]interface{})["stack"], "test stack")
	})

	t.Run("evaluate date", func(t *testing.T) {
		val, err := page.Evaluate(`() => ({ date: new Date("2020-05-27T01:31:38.506Z") })`)
		require.NoError(t, err)
		require.Equal(t, map[string]interface{}{"date": time.Date(2020, 5, 27, 1, 31, 38, 506000000, time.UTC)}, val)
	})

	t.Run("roundtrip date", func(t *testing.T) {
		date := time.Date(2020, 5, 27, 1, 31, 38, 506000000, time.Local)
		val, err := page.Evaluate(`date => date`, date)
		require.NoError(t, err)
		// val is time.Time, but timezone is UTC
		require.True(t, date.Equal(val.(time.Time)))
	})

	t.Run("should evaluate url", func(t *testing.T) {
		val, err := page.Evaluate(`() => ({ someKey: new URL('https://user:pass@example.com/?foo=bar#hi') })`)
		require.NoError(t, err)
		u, ok := val.(map[string]interface{})["someKey"].(*url.URL)
		require.True(t, ok)
		require.Equal(t, "https", u.Scheme)
		require.Equal(t, "user", u.User.Username())
		require.Equal(t, "example.com", u.Host)
		require.Equal(t, "/", u.Path)
		require.Equal(t, "foo=bar", u.RawQuery)
		require.Equal(t, "hi", u.Fragment)
	})

	t.Run("roundtrip url", func(t *testing.T) {
		u, err := url.Parse("https://user:pass@example.com/?foo=bar#hi")
		require.NoError(t, err)
		val, err := page.Evaluate(`url => url`, u)
		require.NoError(t, err)
		require.Equal(t, u, val.(*url.URL))
	})

	t.Run("roundtrip complex url", func(t *testing.T) {
		u, err := url.Parse("https://user:password@www.contoso.com:80/Home/Index.htm?q1=v1&q2=v2#FragmentName")
		require.NoError(t, err)
		val, err := page.Evaluate(`url => url`, u)
		require.NoError(t, err)
		require.Equal(t, u, val.(*url.URL))
	})

	t.Run("jsonvalue url", func(t *testing.T) {
		u, err := url.Parse("https://example.com/")
		require.NoError(t, err)
		val, err := page.Evaluate(`() => ({ someKey: new URL("https://example.com/") })`)
		require.NoError(t, err)
		require.Equal(t, u, val.(map[string]interface{})["someKey"].(*url.URL))
	})

	t.Run("support float64", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, 42.5)
		require.NoError(t, err)
		require.Equal(t, 42.5, val)
	})

	t.Run("unkown as undefined", func(t *testing.T) {
		val, err := page.Evaluate(`a => a`, struct{}{})
		require.NoError(t, err)
		require.Equal(t, nil, val)
	})
}

func TestEvaluateTransferTypedArrays(t *testing.T) {
	BeforeEach(t)

	testTypedArray := func(t *testing.T, typedArray string, expected []float64, valueSuffix string) {
		t.Run(typedArray, func(t *testing.T) {
			val, err := page.Evaluate(fmt.Sprintf(`() => new %s([1%s, 2%s, 3%s])`, typedArray, valueSuffix, valueSuffix, valueSuffix))
			require.NoError(t, err)
			require.Equal(t, expected, val.([]float64))
		})
	}

	testTypedArray(t, "Int8Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Uint8Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Uint8ClampedArray", []float64{1, 2, 3}, "")
	testTypedArray(t, "Int16Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Uint16Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Int32Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Uint32Array", []float64{1, 2, 3}, "")
	testTypedArray(t, "Float32Array", []float64{1.5, 2.5, 3.5}, ".5")
	testTypedArray(t, "Float64Array", []float64{1.5, 2.5, 3.5}, ".5")
	testTypedArray(t, "BigInt64Array", []float64{1, 2, 3}, "n")
	testTypedArray(t, "BigUint64Array", []float64{1, 2, 3}, "n")
}
