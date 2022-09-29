package endpoint

import (
	"net/http"
	"strings"

	"github.com/pindamonhangaba/apiculi/quick_schema"

	"github.com/labstack/echo/v4"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

func Echo[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIRouteDescriber, next Endpoint[C, P, Q, B, D]) (string, string, echo.HandlerFunc) {

	fillOpenAPIRoute[C, P, Q, B, D](endpointPath{
		verb: p.verb,
		path: routerPathToOpenAPIPath(p.path),
	}, d)

	return string(p.verb), p.path, func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		cc, _ := user.Claims.(C)

		m := map[string]string{}
		psch := quick_schema.GetSchema[P]()
		if psch != nil {
			for _, p := range psch.Children {
				m[p.Name] = c.Param(p.Name)
			}
		}
		prs, err := mapToStruct(m, *new(P))
		if err != nil {
			return errors.Wrap(err, "params")
		}

		m = map[string]string{}
		for k, v := range c.Request().URL.Query() {
			m[k] = strings.Join(v, ",")
		}
		q, err := mapToStruct(m, *new(Q))
		if err != nil {
			return errors.Wrap(err, "query")
		}

		b := new(B)
		if has([]httpVerb{PUT, POST, DELETE, PATCH}, p.verb) {
			err = c.Bind(b)
			if err != nil {
				return errors.Wrap(err, "body")
			}
		}

		input := EndpointInput[C, P, Q, B]{
			Claims: cc,
			Params: prs,
			Query:  q,
			Body:   *b,
		}

		r, err := next(input)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, r)
	}
}
