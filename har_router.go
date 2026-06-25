package playwright

import (
	"errors"
)

type harRouter struct {
	localUtils     *localUtilsImpl
	harId          string
	notFoundAction HarNotFound
	urlOrPredicate any
	err            error
}

func (r *harRouter) addContextRoute(context BrowserContext) error {
	if r.err != nil {
		return r.err
	}
	err := context.Route(r.urlOrPredicate, func(route Route) {
		err := r.handle(route)
		if err != nil {
			logger.Error("Error handling context route", "error", err)
		}
	})
	if err != nil {
		return err
	}
	return r.err
}

func (r *harRouter) addPageRoute(page Page) error {
	if r.err != nil {
		return r.err
	}
	err := page.Route(r.urlOrPredicate, func(route Route) {
		err := r.handle(route)
		if err != nil {
			logger.Error("Error handling page route", "error", err)
		}
	})
	if err != nil {
		return err
	}
	return r.err
}

func (r *harRouter) dispose() {
	go r.localUtils.HarClose(r.harId) //nolint:errcheck
}

func (r *harRouter) handle(route Route) error {
	if r.err != nil {
		return r.err
	}
	request := route.Request()
	postData, err := request.PostDataBuffer()
	if err != nil {
		return err
	}
	// Use the actual on-wire headers (ordered, with duplicates) for matching,
	// mirroring upstream which passes request.headersArray().
	headers, err := request.HeadersArray()
	if err != nil {
		return err
	}
	lookup := harLookupOptions{
		HarId:               r.harId,
		URL:                 request.URL(),
		Method:              request.Method(),
		Headers:             headers,
		IsNavigationRequest: request.IsNavigationRequest(),
	}
	// Omit postData for body-less requests; an empty buffer would otherwise be
	// sent as "" and fail to match HAR entries that have no recorded body.
	if len(postData) > 0 {
		lookup.PostData = postData
	}
	response, err := r.localUtils.HarLookup(lookup)
	if err != nil {
		return err
	}
	switch response.Action {
	case "redirect":
		if response.RedirectURL == nil {
			return errors.New("redirect url is null")
		}
		return route.(*routeImpl).redirectedNavigationRequest(*response.RedirectURL)
	case "fulfill":
		// If the response status is -1, the request was canceled or stalled, so we just stall it here.
		// See https://github.com/microsoft/playwright/issues/29311.
		if response.Status != nil && *response.Status == -1 {
			return nil
		}
		if response.Body == nil {
			return errors.New("fulfill body is null")
		}
		return route.Fulfill(RouteFulfillOptions{
			Body:   *response.Body,
			Status: response.Status,
			// route.Fulfill does not support multiple set-cookie headers, so we merge them into one.
			Headers: mergeHeaders(response.Headers),
		})
	case "error":
		if response.Message != nil {
			logger.Error("har action error", "error", *response.Message)
		}
		fallthrough
	case "noentry":
	}
	if r.notFoundAction == *HarNotFoundAbort {
		return route.Abort()
	}
	return route.Fallback()
}

func newHarRouter(localUtils *localUtilsImpl, file string, notFoundAction HarNotFound, urlOrPredicate any) *harRouter {
	harId, err := localUtils.HarOpen(file)
	var url any = "**/*"
	if urlOrPredicate != nil {
		url = urlOrPredicate
	}
	return &harRouter{
		localUtils:     localUtils,
		harId:          harId,
		notFoundAction: notFoundAction,
		urlOrPredicate: url,
		err:            err,
	}
}
