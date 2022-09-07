package endpoint

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/pindamonhangaba/apiculi/quick_schema"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pkg/errors"
)

// Based on Google JSONC styleguide
// https://google.github.io/styleguide/jsoncstyleguide.xml

type errorResponse struct {
	Error generalError `json:"error"`
}

type generalError struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Errors  []detailError `json:"errors,omitempty"`
}

type detailError struct {
	Domain       string  `json:"domain"`
	Reason       string  `json:"reason"`
	Message      string  `json:"message"`
	Location     *string `json:"location,omitempty"`
	LocationType *string `json:"locationType,omitempty"`
	ExtendedHelp *string `json:"extendedHelp,omitempty"`
	SendReport   *string `json:"sendReport,omitempty"`
}

type DataResponse[T dataer] struct {
	// Client sets this value and server echos data in the response
	Context string `json:"context,omitempty"`
	Data    T      `json:"data"`
}

type dataer interface {
	data()
}

type DataDetail struct {
	// The kind property serves as a guide to what type of information this particular object stores
	Kind string `json:"kind" example:"resource"`
	// Indicates the language of the rest of the properties in this object (BCP 47)
	Language string `json:"lang,omitempty" example:"pt-br"`
}
type CollectionDetail struct {
	// The number of items in this result set
	CurrentItemCount int64 `json:"currentItemCount" example:"1"`
	// The number of items in the result
	ItemsPerPage int64 `json:"itemsPerPage" example:"10"`
	// The index of the first item in data.items
	StartIndex int64 `json:"startIndex" example:"1"`
	// The total number of items available in this set
	TotalItems int64 `json:"totalItems" example:"100"`
	// The index of the current page of items
	PageIndex int64 `json:"pageIndex" example:"1"`
	// The total number of pages in the result set.
	TotalPages int64 `json:"totalPages" example:"10"`
}

func (d DataDetail) Data() {}

type SingleItemData[T any] struct {
	DataDetail
	Item T `json:"item"`
}

func (d SingleItemData[T]) data() {}

type CollectionItemData[T any] struct {
	DataDetail
	Items []T `json:"items"`
	CollectionDetail
}

func (d CollectionItemData[T]) data() {}

type EndpointInput[C, P, Q, B any] struct {
	Claims C
	Params P
	Query  Q
	Body   B
}

func mapToStruct[K any, M ~map[string]K, S any](in M, out S) (S, error) {
	params, err := json.Marshal(in)
	if err != nil {
		return out, err
	}

	err = json.Unmarshal((params), &out)
	if err != nil {
		return out, err
	}
	return out, err
}

type RouteDescription struct {
	Summary     string
	Description string
}

func describe(a, b string) RouteDescription {
	return RouteDescription{
		Summary:     a,
		Description: b,
	}
}

func route(r string) string {
	return r
}

var nonWordRgx = regexp.MustCompile(`\W`)
var underlinesRgx = regexp.MustCompile(`_`)

var unnamedCount = 0

func makeNameFromRoute(s string) string {
	r := strings.Split(s, "/")

	for i := range r {
		idx := len(r) - 1 - i
		if strings.Count(r[idx], ":") == 0 {
			return underlinesRgx.ReplaceAllString(nonWordRgx.ReplaceAllString(r[idx], "_"), "_")
		}
	}
	sufx := strconv.Itoa(unnamedCount)
	if unnamedCount == 0 {
		sufx = ""
	}
	return "unnamed" + sufx
}

type Endpoint[C, P, Q, B any, D dataer] func(EndpointInput[C, P, Q, B]) DataResponse[D]

type OpenAPIDescriber func(func(string, string, *openapi3.T))

type OpenAPI struct {
	t openapi3.T
}

func (op *OpenAPI) Describe(title, description string) OpenAPIDescriber {
	return func(f func(string, string, *openapi3.T)) {
		f(title, description, &op.t)
	}
}

func (op *OpenAPI) T() *openapi3.T {
	return &op.t
}

func (op *OpenAPI) AddServer(url, description string) *openapi3.T {
	if op.t.Servers == nil {
		op.t.Servers = openapi3.Servers{}
	}
	op.t.Servers = append(op.t.Servers, &openapi3.Server{
		URL:         url,
		Description: description,
	})
	return &op.t
}

func NewOpenAPI(title, version string) OpenAPI {
	return OpenAPI{
		t: openapi3.T{
			OpenAPI: "3.0",
			Info: &openapi3.Info{
				Title:   title,
				Version: version,
			},
		},
	}
}

type endpointDescription [2]string

type endpointPath struct {
	verb httpVerb
	path string
}

type httpVerb string

const (
	GET    httpVerb = "GET"
	POST   httpVerb = "POST"
	PUT    httpVerb = "PUT"
	PATCH  httpVerb = "PATCH"
	DELETE httpVerb = "DELETE"
)

func Get(path string) endpointPath {
	return endpointPath{GET, path}
}
func Post(path string) endpointPath {
	return endpointPath{POST, path}
}
func Put(path string) endpointPath {
	return endpointPath{PUT, path}
}
func Patch(path string) endpointPath {
	return endpointPath{PATCH, path}
}
func Delete(path string) endpointPath {
	return endpointPath{DELETE, path}
}

func fillOpenAPIRoute[C, P, Q, B any, D dataer](p endpointPath, d OpenAPIDescriber) {
	d(func(tit, desc string, swag *openapi3.T) {
		prepo, err := makeParams[P]("path")
		if err != nil {
			panic(errors.Wrap(err, "bad api data"))
		}
		params := openapi3.Parameters{}
		for param, pv := range prepo {
			err := validatePathParamVar(p.path, param)
			if err != nil {
				panic(errors.Wrap(err, "bad param data"))
			}
			pv.Ref = ""
			params = append(params, pv)
		}

		prepo, err = makeParams[Q]("query")
		if err != nil {
			panic(errors.Wrap(err, "bad api data"))
		}
		for _, pv := range prepo {
			pv.Ref = ""
			params = append(params, pv)
		}

		bodyTypeNodeSchema := quick_schema.GetSchema[B]()
		bodySchema, bodySchemarepo := buildSchemaRepo(*bodyTypeNodeSchema)
		requestBody := &openapi3.RequestBody{
			Description: "Request data",
			Content:     openapi3.NewContentWithJSONSchema(bodySchema),
		}

		responseNodeSchema := quick_schema.GetSchema[DataResponse[D]]()
		responseSchema, responseSchemarepo := buildSchemaRepo(*responseNodeSchema)
		response := &openapi3.Response{
			Description: nil,
			Content:     openapi3.NewContentWithJSONSchema(responseSchema),
		}

		if swag.Components.Schemas == nil {
			swag.Components.Schemas = openapi3.Schemas{}
		}

		for n, val := range bodySchemarepo {
			if val == nil {
				panic("unexpected nil bodySchema")
			}
			swag.Components.Schemas[n] = openapi3.NewSchemaRef("", val)
		}
		for n, val := range responseSchemarepo {
			if val == nil {
				panic("unexpected nil responseSchema")
			}
			swag.Components.Schemas[n] = openapi3.NewSchemaRef("", val)
		}

		op := &openapi3.Operation{
			Parameters: params,
			RequestBody: &openapi3.RequestBodyRef{
				//Ref:   "#/components/requestBodies/someRequestBody",
				Value: requestBody,
			},
			Responses: openapi3.Responses{
				"200": &openapi3.ResponseRef{
					//Ref:   "#/components/responses/someResponse",
					Value: response,
				},
			},
		}
		pitem := &openapi3.PathItem{}
		if swag.Paths == nil {
			swag.Paths = openapi3.Paths{}
		}
		if swag.Paths[p.path] != nil {
			pitem = swag.Paths[p.path]
		}
		switch p.verb {
		case GET:
			pitem.Get = op
		case POST:
			pitem.Post = op
		case PATCH:
			pitem.Patch = op
		case PUT:
			pitem.Put = op
		case DELETE:
			pitem.Delete = op
		}
		swag.Paths[p.path] = pitem
	})
}

func makeParams[T any](in string) (map[string]*openapi3.ParameterRef, error) {
	n := quick_schema.GetSchema[T]()
	if n == nil {
		return nil, nil
	}
	sch, repo := buildSchemaRepo(*n)

	if sch.Type != "object" {
		return nil, errors.New("parameter's type must be a object")
	}

	params := []*openapi3.Parameter{}
	for pname, p := range sch.Properties {
		required := has(sch.Required, pname)

		pram := &openapi3.Parameter{
			Description: p.Value.Description,
			Name:        pname,
			In:          in,
			Required:    required,
			Schema:      p,
		}
		for n, r := range repo {
			if n == p.Value.Title {
				pram.Schema = &openapi3.SchemaRef{
					Ref:   "#/components/schemas/" + n,
					Value: r,
				}
			}
		}
		params = append(params, pram)
	}

	prepo := map[string]*openapi3.ParameterRef{}
	for _, p := range params {
		prepo[p.Name] = &openapi3.ParameterRef{
			Ref:   "#/components/parameters/" + p.Name,
			Value: p,
		}
	}
	return prepo, nil
}

func makeRefableName(pkg, typ string) string {
	s := strings.Split(pkg, "/")
	return strings.ReplaceAll(s[len(s)-1], ".", "_") + typ
}

func buildSchemaRepo(n quick_schema.Node) (*openapi3.Schema, map[string]*openapi3.Schema) {
	repo := map[string]*openapi3.Schema{}
	var schemafy func(n quick_schema.Node) *openapi3.Schema
	schemafy = func(n quick_schema.Node) *openapi3.Schema {
		nname := makeRefableName(n.Package, n.Type)
		s := openapi3.NewSchema()
		s.Title = n.Name
		s.Type = n.Format
		s.Format = n.Type

		if s.Type == "object" {
			s.Properties = make(openapi3.Schemas)
			for _, p := range n.Children {
				pname := "#/components/schemas/" + makeRefableName(p.Package, p.Type)
				ps := schemafy(p)
				if p.Format == "pointer" && len(p.Children) == 1 {
					ps = schemafy(p.Children[0])
				} else {
					s.Required = append(s.Required, p.Name)
				}
				if ps.Format != "object" {
					pname = ""
				}
				sr := openapi3.NewSchemaRef(pname, ps)
				s.Properties[p.Name] = sr
			}
			_, ok := repo[nname]
			if !ok {
				repo[nname] = s
			}
		} else if s.Type == "array" && len(n.Children) == 1 {
			ps := schemafy(n.Children[0])
			name := ""
			if ps.Type == "object" || ps.Type == "array" {
				name = "#/components/schemas/" + makeRefableName(n.Children[0].Package, n.Children[0].Type)
			}
			s.Items = openapi3.NewSchemaRef(name, ps)
		}

		return s
	}
	sch := schemafy(n)
	return sch, repo
}

func has[T comparable](hs []T, n T) bool {
	for _, v := range hs {
		if v == n {
			return true
		}
	}
	return false
}

func validatePathParamVar(path, param string) error {
	if !strings.Contains(path, "{"+param+"}") {
		return errors.Errorf("declared path parameter \"%s\" needs to be defined as a path parameter in \"%s\" ", param, path)
	}
	return nil
}
