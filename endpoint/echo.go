package endpoint

import (
	"net/http"
	"strings"

	"github.com/pindamonhangaba/apiculi/quick_schema"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

func EchoWithContext[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIRouteDescriber, next EndpointWithContext[C, P, Q, B, D, echo.Context]) (string, string, echo.HandlerFunc) {
	fillOpenAPIRoute[C, P, Q, B, D](endpointPath{
		verb: p.verb,
		path: routerPathToOpenAPIPath(p.path),
	}, d)

	return string(p.verb), p.path, func(c echo.Context) error {

		cc, prs, q, b, err := parseBodyEcho[C, P, Q, B, D](p, c)

		input := EndpointInput[C, P, Q, B]{
			Claims: cc,
			Params: prs,
			Query:  q,
		}
		if b != nil {
			input.Body = *b
		}

		r, err := next(input, c)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, r)
	}
}

func Echo[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIRouteDescriber, next Endpoint[C, P, Q, B, D]) (string, string, echo.HandlerFunc) {

	fillOpenAPIRoute[C, P, Q, B, D](endpointPath{
		verb: p.verb,
		path: routerPathToOpenAPIPath(p.path),
	}, d)

	return string(p.verb), p.path, func(c echo.Context) error {

		cc, prs, q, b, err := parseBodyEcho[C, P, Q, B, D](p, c)

		input := EndpointInput[C, P, Q, B]{
			Claims: cc,
			Params: prs,
			Query:  q,
		}
		if b != nil {
			input.Body = *b
		}

		r, err := next(input)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, r)
	}
}

func parseBodyEcho[C, P, Q, B any, D dataer](p endpointPath, c echo.Context) (cc C, prs P, q Q, b *B, err error) {
	switch c.Request().Header.Get("Content-Type") {
	case "application/json", "application/x-www-form-urlencoded", "multipart/form-data":
	default:
		return cc, prs, q, b, errors.Errorf(`unsupported content-type %s, must be "application/json" or "application/x-www-form-urlencoded"`)
	}

	// ignore claims if type is "any"
	if any(*new(C)) != nil {
		user, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return cc, prs, q, b, errors.Errorf("unexpected claims type")
		}
		cc, _ = user.Claims.(C)
	}

	m := map[string]string{}
	psch := quick_schema.GetSchema[P]()
	if psch != nil {
		for _, p := range psch.Children {
			m[p.Name] = c.Param(p.Name)
		}
	}
	prs, err = mapToStruct(m, *new(P))
	if err != nil {
		return cc, prs, q, b, errors.Wrap(err, "params")
	}

	m = map[string]string{}
	for k, v := range c.Request().URL.Query() {
		m[k] = strings.Join(v, ",")
	}
	q, err = mapToStruct(m, *new(Q))
	if err != nil {
		return cc, prs, q, b, errors.Wrap(err, "query")
	}

	b = new(B)
	if has([]httpVerb{PUT, POST, DELETE, PATCH}, p.verb) {
		err = c.Bind(b)
		if err != nil {
			return cc, prs, q, b, errors.Wrap(err, "body")
		}
	}

	return cc, prs, q, b, nil
}
