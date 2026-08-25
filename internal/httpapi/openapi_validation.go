package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"sync"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/google/uuid"
)

var registerOpenAPIFormats sync.Once

func newOpenAPIRouter() routers.Router {
	registerOpenAPIFormats.Do(func() {
		openapi3.DefineStringFormatValidator("uuid", openapi3.NewCallbackValidator(func(value string) error {
			_, err := uuid.Parse(value)
			return err
		}))
		openapi3.DefineStringFormatValidator("email", openapi3.NewRegexpFormatValidator(openapi3.FormatOfStringForEmail))
		openapi3.DefineStringFormatValidator("uri-reference", openapi3.NewCallbackValidator(func(value string) error {
			_, err := url.Parse(value)
			return err
		}))
	})
	spec, err := contract.GetSwagger()
	if err != nil {
		panic("load embedded OpenAPI contract: " + err.Error())
	}
	router, err := legacy.NewRouter(spec)
	if err != nil {
		panic("build OpenAPI router: " + err.Error())
	}
	return router
}

func validateOpenAPIRequests(router routers.Router, next http.Handler) http.Handler {
	options := &openapi3filter.Options{
		AuthenticationFunc:  openapi3filter.NoopAuthenticationFunc,
		SkipSettingDefaults: true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, pathParams, err := router.FindRoute(r)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options:    options,
		}
		if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
			writeProblem(w, http.StatusBadRequest, "request is invalid")
			return
		}
		if err := validateRequestFormats(r, route, pathParams); err != nil {
			writeProblem(w, http.StatusBadRequest, "request is invalid")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validateRequestFormats(r *http.Request, route *routers.Route, pathParams map[string]string) error {
	parameters := append(route.PathItem.Parameters, route.Operation.Parameters...)
	for _, parameterRef := range parameters {
		parameter := parameterRef.Value
		if parameter == nil || parameter.Schema == nil || parameter.Schema.Value == nil || parameter.Schema.Value.Format == "" {
			continue
		}
		var value string
		switch parameter.In {
		case openapi3.ParameterInPath:
			value = pathParams[parameter.Name]
		case openapi3.ParameterInQuery:
			value = r.URL.Query().Get(parameter.Name)
		case openapi3.ParameterInHeader:
			value = r.Header.Get(parameter.Name)
		}
		if value != "" {
			parsedValue, err := parseParameterValue(parameter.Schema.Value, value)
			if err != nil {
				return fmt.Errorf("parse %s value: %w", parameter.Name, err)
			}
			if err := parameter.Schema.Value.VisitJSON(parsedValue, openapi3.EnableFormatValidation()); err != nil {
				return fmt.Errorf("validate %s format: %w", parameter.Name, err)
			}
		}
	}

	requestBody := route.Operation.RequestBody
	if requestBody == nil || requestBody.Value == nil || r.Body == nil {
		return nil
	}
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	mediaType := requestBody.Value.Content.Get(contentType)
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return nil
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewReader(payload))
	if len(payload) == 0 {
		return nil
	}
	var document any
	if err := json.Unmarshal(payload, &document); err != nil {
		return err
	}
	return mediaType.Schema.Value.VisitJSON(document, openapi3.VisitAsRequest(), openapi3.EnableFormatValidation())
}

func parseParameterValue(schema *openapi3.Schema, value string) (any, error) {
	if schema.Type.Permits(openapi3.TypeString) {
		return value, nil
	}

	var parsed any
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}
