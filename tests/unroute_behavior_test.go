package playwright_test

import (
	"errors"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

func TestContextUnrouteShouldNotWaitForPendingHandlersToComplete(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, context.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		secondHandlerCalled = true
		err := route.Continue()
		require.NoError(t, err)
	}))

	routeChan := make(chan playwright.Route, 1)
	routeBarrier := make(chan struct{}, 1)

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		require.NoError(t, route.Fallback())
	}

	require.NoError(t, context.Route(regexp.MustCompile(".*"), handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()

	<-routeChan
	require.NoError(t, context.Unroute(regexp.MustCompile(".*"), handler2))

	routeBarrier <- struct{}{}
	wg.Wait()
	require.True(t, secondHandlerCalled)
}

func TestContextUnrouteAllRemovesAllHandlers(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, context.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Abort())
	}))

	require.NoError(t, context.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Abort())
	}))

	require.NoError(t, context.UnrouteAll())
	_, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
}

func TestContextUnrouteAllShouldNotWaitForPendingHandlersToComplete(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, context.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		secondHandlerCalled = true
		err := route.Abort()
		require.NoError(t, err)
	}))

	routeChan := make(chan playwright.Route)
	routeBarrier := make(chan struct{})

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		require.NoError(t, route.Fallback())
	}

	require.NoError(t, context.Route(regexp.MustCompile(".*"), handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan

	didUnroute := false
	wgUnroute := &sync.WaitGroup{}
	wgUnroute.Add(1)
	go func() {
		require.NoError(t, context.UnrouteAll(playwright.BrowserContextUnrouteAllOptions{
			Behavior: playwright.UnrouteBehaviorWait,
		}))
		didUnroute = true
		wgUnroute.Done()
	}()
	time.Sleep(500 * time.Millisecond)
	require.False(t, didUnroute)
	routeBarrier <- struct{}{}
	wgUnroute.Wait()
	require.True(t, didUnroute)
	wg.Wait()
	require.False(t, secondHandlerCalled)
}

func TestContextUnrouteAllShouldNotWaitForPendingHandlersToCompleteIfBehaviorIsIgnoreErrors(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, context.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		secondHandlerCalled = true
		err := route.Abort()
		require.NoError(t, err)
	}))

	routeChan := make(chan playwright.Route)
	routeBarrier := make(chan struct{})

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		panic(errors.New("Handler error"))
	}

	require.NoError(t, context.Route(regexp.MustCompile(".*"), handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan

	didUnroute := false
	wgUnroute := &sync.WaitGroup{}
	wgUnroute.Add(1)
	go func() {
		require.NoError(t, context.UnrouteAll(playwright.BrowserContextUnrouteAllOptions{
			Behavior: playwright.UnrouteBehaviorIgnoreErrors,
		}))
		didUnroute = true
		wgUnroute.Done()
	}()
	time.Sleep(500 * time.Millisecond)
	routeBarrier <- struct{}{}
	wgUnroute.Wait()
	require.True(t, didUnroute)
	wg.Wait()
	require.False(t, secondHandlerCalled)
}

func TestPageCloseShouldNotWaitForActiveRouteHandlersOnTheOwningContext(t *testing.T) {
	BeforeEach(t)

	routeChan := make(chan playwright.Route)
	require.NoError(t, context.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		routeChan <- route
	}))
	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		require.NoError(t, route.Fallback())
	}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan
	require.NoError(t, page.Close())
	wg.Wait()
}

func TestContextCloseShouldNotWaitForActiveRouteHandlersOnTheOwnedPages(t *testing.T) {
	BeforeEach(t)

	routeChan := make(chan playwright.Route)
	require.NoError(t, context.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		routeChan <- route
	}))
	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		require.NoError(t, route.Fallback())
	}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan
	require.NoError(t, context.Close())
	wg.Wait()
}

func TestPageUnrouteShouldNotWaitForPendingHandlersToComplete(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		secondHandlerCalled = true
		require.NoError(t, route.Continue())
	}))

	routeChan := make(chan playwright.Route)
	routeBarrier := make(chan struct{})

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		require.NoError(t, route.Fallback())
	}

	require.NoError(t, page.Route(regexp.MustCompile(".*"), handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan

	require.NoError(t, page.Unroute(regexp.MustCompile(".*"), handler2))
	routeBarrier <- struct{}{}
	wg.Wait()
	require.True(t, secondHandlerCalled)
}

func TestPageUnrouteAllRemovesAllRoutes(t *testing.T) {
	BeforeEach(t)

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		require.NoError(t, route.Abort())
	}))
	require.NoError(t, page.Route("**/empty.html", func(route playwright.Route) {
		require.NoError(t, route.Abort())
	}))

	require.NoError(t, page.UnrouteAll())

	response, err := page.Goto(server.EMPTY_PAGE)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.True(t, response.Ok())
}

func TestPageUnrouteShouldWaitForPendingHandlersToComplete(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		secondHandlerCalled = true
		require.NoError(t, route.Abort())
	}))

	routeChan := make(chan playwright.Route)
	routeBarrier := make(chan struct{})

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		require.NoError(t, route.Fallback())
	}

	require.NoError(t, page.Route("**/*", handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan

	didUnroute := false
	wgUnroute := &sync.WaitGroup{}
	wgUnroute.Add(1)
	go func() {
		require.NoError(t, page.UnrouteAll(playwright.PageUnrouteAllOptions{
			Behavior: playwright.UnrouteBehaviorWait,
		}))
		didUnroute = true
		wgUnroute.Done()
	}()
	time.Sleep(500 * time.Millisecond)
	routeBarrier <- struct{}{}
	wgUnroute.Wait()
	require.True(t, didUnroute)
	wg.Wait()
	require.False(t, secondHandlerCalled)
}

func TestPageUnrouteAllShouldNotWaitForPendingHandlersToCompleteIfBehaviorIsIgnoreErrors(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		secondHandlerCalled = true
		require.NoError(t, route.Abort())
	}))

	routeChan := make(chan playwright.Route)
	routeBarrier := make(chan struct{})

	handler2 := func(route playwright.Route) {
		routeChan <- route
		<-routeBarrier
		panic(errors.New("Handler error"))
	}

	require.NoError(t, page.Route(regexp.MustCompile(".*"), handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan

	didUnroute := false
	wgUnroute := &sync.WaitGroup{}
	wgUnroute.Add(1)
	go func() {
		require.NoError(t, page.UnrouteAll(playwright.PageUnrouteAllOptions{
			Behavior: playwright.UnrouteBehaviorIgnoreErrors,
		}))
		didUnroute = true
		wgUnroute.Done()
	}()
	time.Sleep(500 * time.Millisecond)
	wgUnroute.Wait()
	require.True(t, didUnroute)
	routeBarrier <- struct{}{}
	wg.Wait()
	require.False(t, secondHandlerCalled)
}

func TestPageCloseDoesNotWaitForActiveRouteHandlers(t *testing.T) {
	BeforeEach(t)

	secondHandlerCalled := false

	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		secondHandlerCalled = true
	}))

	routeChan := make(chan playwright.Route)

	handler2 := func(route playwright.Route) {
		routeChan <- route
	}
	require.NoError(t, page.Route("**/*", handler2))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	<-routeChan
	require.NoError(t, page.Close())
	time.Sleep(500 * time.Millisecond)
	require.False(t, secondHandlerCalled)
}

func TestRouteContinueShouldNotThrowIfPageHasBeenClosed(t *testing.T) {
	BeforeEach(t)

	routeChan := make(chan playwright.Route)
	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		routeChan <- route
	}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	route := <-routeChan
	require.NoError(t, page.Close())
	// Should not throw (upstream).
	require.NoError(t, route.Continue())
}

func TestRouteFallbackShouldNotThroIfPageHasbeenClosed(t *testing.T) {
	BeforeEach(t)

	routeChan := make(chan playwright.Route)
	require.NoError(t, page.Route(regexp.MustCompile(".*"), func(route playwright.Route) {
		routeChan <- route
	}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	route := <-routeChan
	require.NoError(t, page.Close())
	// Should not throw (upstream).
	require.NoError(t, route.Fallback())
}

func TestRouteFulfillShouldNotThrowIfPageHasBeenClosed(t *testing.T) {
	BeforeEach(t)

	routeChan := make(chan playwright.Route)
	require.NoError(t, page.Route("**/*", func(route playwright.Route) {
		routeChan <- route
	}))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		_, _ = page.Goto(server.EMPTY_PAGE)
		wg.Done()
	}()
	route := <-routeChan
	require.NoError(t, page.Close())
	// Should not throw (upstream).
	require.NoError(t, route.Fulfill())
}
