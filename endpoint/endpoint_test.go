package endpoint

import (
	"strings"
	"testing"

	"github.com/pindamonhangaba/apiculi/quick_schema"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/andreyvit/diff"
	"github.com/ghodss/yaml"
	"github.com/gofiber/fiber/v2"
)

type TT = EndpointInput[struct {
	Name string `json:"name"`
}, interface{}, struct {
	Context string `json:"context"`
	Number  string `json:"number"`
}, struct {
	A string `json:"a"`
	B string `json:"b"`
}]

func TestFiber(t *testing.T) {

	oapi := NewOpenAPI("Endpoint Docs", "v1.0.1")

	app := fiber.New()

	app.Add(Fiber(
		Get("/api/endpoint/:id"),
		oapi.Describe("List one resource", `
			* Nice list
			* Description ~is here~
		`),
		func(in TT) (DataResponse[SingleItemData[struct {
			A string `json:"a"`
			B string `json:"b"`
		}]], error) {
			return DataResponse[SingleItemData[struct {
				A string `json:"a"`
				B string `json:"b"`
			}]]{
				Context: in.Query.Context,
				Data: SingleItemData[struct {
					A string `json:"a"`
					B string `json:"b"`
				}]{
					DataDetail: DataDetail{
						Kind: "resource",
					},
					Item: in.Body,
				},
			}, nil
		},
	))
}

func TestParams(t *testing.T) {

	expectedYAML := "/query-param-test:\n  post:\n    parameters:\n    - in: query\n      name: AnotherValue\n      required: true\n      schema:\n        items:\n          format: int\n          type: number\n        title: AnotherValue\n        type: array\n    - in: query\n      name: Props\n      required: true\n      schema:\n        format: testParam\n        properties:\n          ParamProp:\n            format: string\n            title: ParamProp\n            type: string\n        required:\n        - ParamProp\n        title: Props\n        type: object\n    - in: query\n      name: SomeValue\n      required: true\n      schema:\n        format: string\n        title: SomeValue\n        type: string\n    responses: null"
	type testParam struct {
		ParamProp string
	}
	type param struct {
		SomeValue    string
		AnotherValue []int
		Props        testParam
	}
	r, err := makeParams[param]("query")
	if err != nil {
		t.Error(err)
	}
	params := openapi3.Parameters{}
	prepo := map[string]*openapi3.ParameterRef{}
	for k, p := range r {
		params = append(params, p)
		p.Ref = ""
		prepo[k] = p
	}
	paths := openapi3.Paths{
		"/query-param-test": &openapi3.PathItem{
			Post: &openapi3.Operation{
				Parameters: params,
			},
		},
	}
	outputYAML, err := yaml.Marshal(paths)
	if err != nil {
		t.Error(err)
	}
	if a, e := strings.TrimSpace(string(outputYAML)), strings.TrimSpace(expectedYAML); a != e {
		t.Errorf("result not as expected:\n%v", diff.LineDiff(e, a))
	}
	t.Log(string(outputYAML))
}

func TestBuildSchemaRepo(t *testing.T) {

	expectedYAML := "schemas:\n  endpointparam:\n    format: param\n    properties:\n      AnotherValue:\n        items:\n          format: int\n          type: number\n        title: AnotherValue\n        type: array\n      Props:\n        format: testParam\n        properties:\n          ParamProp:\n            format: string\n            title: ParamProp\n            type: string\n        required:\n        - ParamProp\n        title: Props\n        type: object\n      SomeValue:\n        format: string\n        title: SomeValue\n        type: string\n    required:\n    - SomeValue\n    - AnotherValue\n    - Props\n    type: object\n  endpointtestParam:\n    format: testParam\n    properties:\n      ParamProp:\n        format: string\n        title: ParamProp\n        type: string\n    required:\n    - ParamProp\n    title: Props\n    type: object"

	type testParam struct {
		ParamProp string
	}
	type param struct {
		SomeValue    string
		AnotherValue []int
		Props        testParam
	}
	n := quick_schema.GetSchema[param]()
	_, repo := buildSchemaRepo(*n)
	schemass := map[string]*openapi3.SchemaRef{}
	for name, schema := range repo {
		schemass[name] = &openapi3.SchemaRef{
			Value: schema,
		}
	}
	components := openapi3.Components{
		Schemas: schemass,
	}
	outputYAML, err := yaml.Marshal(components)
	if err != nil {
		t.Error(err)
	}
	if a, e := strings.TrimSpace(string(outputYAML)), strings.TrimSpace(expectedYAML); a != e {
		t.Errorf("result not as expected:\n%v", diff.LineDiff(e, a))
	}
}
func TestFillOpenAPIRoute(t *testing.T) {

	expectedYAML := "components:\n  schemas:\n    endpointDataDetail:\n      format: DataDetail\n      properties:\n        kind:\n          format: string\n          title: kind\n          type: string\n        lang:\n          format: string\n          title: lang\n          type: string\n      required:\n      - kind\n      - lang\n      title: DataDetail\n      type: object\n    endpointDataResponse[endpoint.SingleItemData[string]]:\n      format: DataResponse[endpoint.SingleItemData[string]]\n      properties:\n        context:\n          format: string\n          title: context\n          type: string\n        data:\n          format: SingleItemData[string]\n          properties:\n            DataDetail:\n              format: DataDetail\n              properties:\n                kind:\n                  format: string\n                  title: kind\n                  type: string\n                lang:\n                  format: string\n                  title: lang\n                  type: string\n              required:\n              - kind\n              - lang\n              title: DataDetail\n              type: object\n            item:\n              format: string\n              title: item\n              type: string\n          required:\n          - DataDetail\n          - item\n          title: data\n          type: object\n      required:\n      - context\n      - data\n      type: object\n    endpointSingleItemData[string]:\n      format: SingleItemData[string]\n      properties:\n        DataDetail:\n          format: DataDetail\n          properties:\n            kind:\n              format: string\n              title: kind\n              type: string\n            lang:\n              format: string\n              title: lang\n              type: string\n          required:\n          - kind\n          - lang\n          title: DataDetail\n          type: object\n        item:\n          format: string\n          title: item\n          type: string\n      required:\n      - DataDetail\n      - item\n      title: data\n      type: object\n    endpointbody:\n      format: body\n      properties:\n        Content:\n          format: string\n          title: Content\n          type: string\n      required:\n      - Content\n      type: object\ninfo:\n  title: Endpoint Docs\n  version: v1.0.1\nopenapi: \"3.0\"\npaths:\n  /api/endpoint/{ParamProp}:\n    get:\n      parameters:\n      - in: path\n        name: ParamProp\n        required: true\n        schema:\n          format: string\n          title: ParamProp\n          type: string\n      - in: query\n        name: SomeValue\n        required: true\n        schema:\n          format: string\n          title: SomeValue\n          type: string\n      - in: query\n        name: AnotherValue\n        required: true\n        schema:\n          items:\n            format: int\n            type: number\n          title: AnotherValue\n          type: array\n      - in: query\n        name: Props\n        required: true\n        schema:\n          format: testParam\n          properties:\n            ParamProp:\n              format: string\n              title: ParamProp\n              type: string\n          required:\n          - ParamProp\n          title: Props\n          type: object\n      requestBody:\n        content:\n          application/json:\n            schema:\n              format: body\n              properties:\n                Content:\n                  format: string\n                  title: Content\n                  type: string\n              required:\n              - Content\n              type: object\n        description: Request data\n      responses:\n        \"200\":\n          content:\n            application/json:\n              schema:\n                format: DataResponse[endpoint.SingleItemData[string]]\n                properties:\n                  context:\n                    format: string\n                    title: context\n                    type: string\n                  data:\n                    format: SingleItemData[string]\n                    properties:\n                      DataDetail:\n                        format: DataDetail\n                        properties:\n                          kind:\n                            format: string\n                            title: kind\n                            type: string\n                          lang:\n                            format: string\n                            title: lang\n                            type: string\n                        required:\n                        - kind\n                        - lang\n                        title: DataDetail\n                        type: object\n                      item:\n                        format: string\n                        title: item\n                        type: string\n                    required:\n                    - DataDetail\n                    - item\n                    title: data\n                    type: object\n                required:\n                - context\n                - data\n                type: object"

	oapi := NewOpenAPI("Endpoint Docs", "v1.0.1")
	type claimed struct {
		UserID string
	}
	type testParam struct {
		ParamProp string
	}
	type param struct {
		SomeValue    string
		AnotherValue []int
		Props        testParam
	}
	type body struct {
		Content string
	}
	fillOpenAPIRoute[claimed, testParam, param, body, SingleItemData[string]](endpointPath{
		path: "/api/endpoint/{ParamProp}",
		verb: GET,
	}, oapi.Describe("title", "description"))

	outputYAML, err := yaml.Marshal(oapi.T())
	if err != nil {
		t.Error(err)
	}
	if a, e := strings.TrimSpace(string(outputYAML)), strings.TrimSpace(expectedYAML); a != e {
		t.Errorf("result not as expected:\n%v", diff.LineDiff(e, a))
	}
}
