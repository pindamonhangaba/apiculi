package endpoint

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/pindamonhangaba/apiculi/quick_schema"
	"github.com/pkg/errors"
)

func writeErrJSON(w http.ResponseWriter, statusCode int, err error) {
	e := errorResponse{
		Error: generalError{
			Message: err.Error(),
		},
	}
	b, err := json.Marshal(e)
	if err != nil {
		b = []byte(err.Error())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func writeJSON[T any](w http.ResponseWriter, statusCode int, data T) {
	b, err := json.Marshal(data)
	if err != nil {
		writeErrJSON(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(b)
}

func Gorilla[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIRouteDescriber, next Endpoint[C, P, Q, B, D]) (string, string, http.HandlerFunc) {

	fillOpenAPIRoute[C, P, Q, B, D](p, d)

	return string(p.verb), p.path, func(w http.ResponseWriter, req *http.Request) {
		user := req.Context().Value("user").(*jwt.Token)
		cc, _ := user.Claims.(C)

		vars := mux.Vars(req)
		m := map[string]string{}
		psch := quick_schema.GetSchema[P]()
		if psch != nil {
			for _, p := range psch.Children {
				m[p.Name] = vars[p.Name]
			}
		}
		prs, err := mapToStruct(m, *new(P))
		if err != nil {
			writeErrJSON(w, http.StatusBadRequest, errors.Wrap(err, "params"))
			return
		}

		if req.Header.Get("Content-Type") != "application/json" {
			writeErrJSON(w, http.StatusBadRequest, errors.New("invalid content type"))
			return
		}

		m = map[string]string{}
		for k, v := range req.URL.Query() {
			m[k] = strings.Join(v, ",")
		}
		q, err := mapToStruct(m, *new(Q))
		if err != nil {
			writeErrJSON(w, http.StatusInternalServerError, err)
			return
		}

		b := new(B)
		if has([]httpVerb{PUT, POST, DELETE, PATCH}, p.verb) {
			err := json.NewDecoder(req.Body).Decode(&b)
			if err != nil {
				writeErrJSON(w, http.StatusBadRequest, errors.Wrap(err, "body"))
				return
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
			writeErrJSON(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, r)
	}
}
