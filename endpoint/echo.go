package endpoint

import (
	"net/http"
	"strings"

	"github.com/pindamonhangaba/apiculi/quick_schema"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

func EchoWithContext[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIRouteDescriber, next EndpointWithContext[C, P, Q, B, D, echo.Context]) (string, string, echo.HandlerFunc) {
	fillOpenAPIRoute[C, P, Q, B, D](endpointPath{
		verb: p.verb,
		path: routerPathToOpenAPIPath(p.path),
	}, d)

	return string(p.verb), p.path, func(c echo.Context) error {

		cc, prs, q, b, err := parseBodyEcho[C, P, Q, B, D](p, c)
		if err != nil {
			return err
		}

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
		if err != nil {
			return err
		}

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
	if _, ok := Find([]string{http.MethodGet, http.MethodConnect, http.MethodHead, http.MethodTrace, http.MethodOptions}, string(p.verb)); !ok {
		contt := strings.Split(c.Request().Header.Get("Content-Type"), ";")[0]
		switch contt {
		case "application/json", "application/x-www-form-urlencoded", "multipart/form-data":
		default:
			return cc, prs, q, b, errors.Errorf(`unsupported content-type %s, must be "application/json" or "application/x-www-form-urlencoded"`, contt)
		}
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

	qMap := map[string]any{}
	for k, v := range c.Request().URL.Query() {
		if len(v) > 1 {
			qMap[k] = v
		} else if len(v) == 1 {
			qMap[k] = v[0]
		}
	}
	q, err = mapToStruct(qMap, *new(Q))
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
