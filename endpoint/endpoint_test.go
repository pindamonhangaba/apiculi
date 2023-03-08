package endpoint

import (
	"encoding/json"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pindamonhangaba/apiculi/quick_schema"
	diff "github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
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

func diffJSON(a, b []byte) (string, error) {
	differ := diff.New()
	d, err := differ.Compare(a, b)
	if err != nil {
		return "", err
	}
	if d.Modified() {
		var aJson map[string]interface{}
		json.Unmarshal(a, &aJson)

		config := formatter.AsciiFormatterConfig{
			ShowArrayIndex: true,
			Coloring:       true,
		}

		formatter := formatter.NewAsciiFormatter(aJson, config)
		diffString, err := formatter.Format(d)
		if err != nil {
			return "", err
		}
		return diffString, nil
	}
	return "", nil
}

func TestFiber(t *testing.T) {

	oapi := NewOpenAPI("Endpoint Docs", "v1.0.1")

	app := fiber.New()

	app.Add(Fiber(
		Get("/api/endpoint/:id"),
		oapi.Route("List one resource", `
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

	expectedJSON := []byte(`{"/query-param-test":{"post":{"parameters":[{"in":"query","name":"SomeValue","required":true,"schema":{"format":"string","title":"SomeValue","type":"string"}},{"in":"query","name":"AnotherValue","required":true,"schema":{"items":{"format":"int","type":"number"},"title":"AnotherValue","type":"array"}},{"in":"query","name":"Props","required":true,"schema":{"format":"testParam","properties":{"ParamProp":{"format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"Props","type":"object"}}],"responses":null}}}`)

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
	j, err := json.Marshal(paths)
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON(expectedJSON, j)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}
}

func TestBuildSchemaRepo(t *testing.T) {

	expectedJSON := []byte(`{"schemas":{"endpointparam":{"format":"param","properties":{"AnotherValue":{"items":{"format":"int","type":"number"},"title":"AnotherValue","type":"array"},"Props":{"format":"testParam","properties":{"ParamProp":{"format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"Props","type":"object"},"SomeValue":{"format":"string","title":"SomeValue","type":"string"}},"required":["SomeValue","AnotherValue","Props"],"type":"object"},"endpointtestParam":{"format":"testParam","properties":{"ParamProp":{"format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"Props","type":"object"}}}`)

	type testParam struct {
		ParamProp string
	}
	type param struct {
		SomeValue    string
		AnotherValue []int
		Props        testParam
	}
	n := quick_schema.GetSchema[param]()
	repo := buildSchemaRepo(*n)
	schemass := map[string]*openapi3.SchemaRef{}
	for name, schema := range repo.Repo {
		schemass[name] = &openapi3.SchemaRef{
			Value: schema,
		}
	}
	components := openapi3.Components{
		Schemas: schemass,
	}
	j, err := components.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON(expectedJSON, j)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}
}

func TestFillOpenAPIRoute(t *testing.T) {

	expectedJSON := []byte(`{"components":{"schemas":{"endpointDataDetail":{"format":"DataDetail","properties":{"kind":{"format":"string","title":"kind","type":"string"},"lang":{"format":"string","title":"lang","type":"string"}},"required":["kind","lang"],"title":"DataDetail","type":"object"},"endpointDataResponse[endpoint.SingleItemData[string]]":{"format":"DataResponse[endpoint.SingleItemData[string]]","properties":{"context":{"format":"string","title":"context","type":"string"},"data":{"format":"SingleItemData[string]","properties":{"DataDetail":{"format":"DataDetail","properties":{"kind":{"format":"string","title":"kind","type":"string"},"lang":{"format":"string","title":"lang","type":"string"}},"required":["kind","lang"],"title":"DataDetail","type":"object"},"item":{"format":"string","title":"item","type":"string"}},"required":["DataDetail","item"],"title":"data","type":"object"}},"required":["context","data"],"type":"object"},"endpointSingleItemData[string]":{"format":"SingleItemData[string]","properties":{"DataDetail":{"format":"DataDetail","properties":{"kind":{"format":"string","title":"kind","type":"string"},"lang":{"format":"string","title":"lang","type":"string"}},"required":["kind","lang"],"title":"DataDetail","type":"object"},"item":{"format":"string","title":"item","type":"string"}},"required":["DataDetail","item"],"title":"data","type":"object"},"endpointbody":{"format":"body","properties":{"Content":{"format":"string","title":"Content","type":"string"}},"required":["Content"],"type":"object"}}},"info":{"title":"Endpoint Docs","version":"v1.0.1"},"openapi":"3.0","paths":{"/api/endpoint/{ParamProp}":{"get":{"parameters":[{"in":"path","name":"ParamProp","required":true,"schema":{"format":"string","title":"ParamProp","type":"string"}},{"in":"query","name":"Props","required":true,"schema":{"format":"testParam","properties":{"ParamProp":{"format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"Props","type":"object"}},{"in":"query","name":"SomeValue","required":true,"schema":{"format":"string","title":"SomeValue","type":"string"}},{"in":"query","name":"AnotherValue","required":true,"schema":{"items":{"format":"int","type":"number"},"title":"AnotherValue","type":"array"}}],"requestBody":{"content":{"application/json":{"schema":{"format":"body","properties":{"Content":{"format":"string","title":"Content","type":"string"}},"required":["Content"],"type":"object"}}},"description":"Request data"},"responses":{"200":{"content":{"application/json":{"schema":{"format":"DataResponse[endpoint.SingleItemData[string]]","properties":{"context":{"format":"string","title":"context","type":"string"},"data":{"format":"SingleItemData[string]","properties":{"DataDetail":{"format":"DataDetail","properties":{"kind":{"format":"string","title":"kind","type":"string"},"lang":{"format":"string","title":"lang","type":"string"}},"required":["kind","lang"],"title":"DataDetail","type":"object"},"item":{"format":"string","title":"item","type":"string"}},"required":["DataDetail","item"],"title":"data","type":"object"}},"required":["context","data"],"type":"object"}}}}}}}}}`)
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
	}, oapi.Route("title", "description"))

	j, err := oapi.T().MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON(expectedJSON, j)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", "d")
	}
}

type PushTokens struct {
	pq.StringArray
}

func TestOptionalOmitempty(t *testing.T) {

	expectedJSON := []byte(`{"schemas":{"github_com_pindamonhangaba_apiculi_endpoint":{"example":"","format":"param","properties":{"AnotherValue":{"example":"","items":{"example":"","format":"int","type":"number"},"title":"AnotherValue","type":"array"},"Props":{"example":"","format":"testParam","properties":{"ParamProp":{"example":"","format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"github_com_pindamonhangaba_apiculi_endpointProps","type":"object"},"PushTokens":{"example":"","format":"StringArray","items":{"example":"","format":"string","type":"string"},"title":"github_com_lib_pqPushTokens","type":"array"},"some_value":{"example":"","format":"string","nullable":true,"title":"some_value","type":"string"}},"required":["AnotherValue","Props","PushTokens"],"title":"github_com_pindamonhangaba_apiculi_endpoint","type":"object"},"github_com_pindamonhangaba_apiculi_endpointProps":{"example":"","format":"testParam","properties":{"ParamProp":{"example":"","format":"string","title":"ParamProp","type":"string"}},"required":["ParamProp"],"title":"github_com_pindamonhangaba_apiculi_endpointProps","type":"object"}}}`)

	type testParam struct {
		ParamProp string
	}
	type param struct {
		SomeValue    string `json:"some_value,omitempty"`
		AnotherValue []int
		Props        testParam
		PushTokens   PushTokens
	}
	n := quick_schema.GetSchema[param]()
	repo := buildSchemaRepo(*n)
	schemass := map[string]*openapi3.SchemaRef{}
	for name, schema := range repo.Repo {
		schemass[name] = &openapi3.SchemaRef{
			Value: schema,
		}
	}
	components := openapi3.Components{
		Schemas: schemass,
	}
	j, err := components.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	d, err := diffJSON(expectedJSON, j)
	if err != nil {
		t.Error(err)
	}
	if len(d) > 0 {
		t.Errorf("result not as expected:\n%v", d)
	}
}
