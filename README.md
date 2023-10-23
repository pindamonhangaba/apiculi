# apiculi
Conveniently type-safe API endpoints

How to use:

	package main

	import (
		"context"
		"encoding/json"
		"net/http"
		"os"
		"os/signal"
		"time"

		jwt "github.com/golang-jwt/jwt/v5"
		echojwt "github.com/labstack/echo-jwt/v4"
		"github.com/labstack/echo/v4"
		"github.com/pindamonhangaba/apiculi/endpoint"
	)

	type Collection struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	type FilterCollection struct {
		Name string `json:"name"`
	}

	type Claims struct {
		UserID string `json:"userID"`
		jwt.RegisteredClaims
	}

	type ContextQ struct {
		Context string `json:"context"`
	}

	func main() {
		e := echo.New()
		gJWT := e.Group("")
		jwtConfig := echojwt.Config{
			NewClaimsFunc: func(c echo.Context) jwt.Claims {
				return new(Claims)
			},
			SigningKey:    []byte("secret"),
			ContextKey:    "user",
			SigningMethod: jwt.SigningMethodHS256.Name,
		}
		gJWT.Use(echojwt.WithConfig(jwtConfig))

		oapi := endpoint.NewOpenAPI("API", "v1")
		oapi.AddServer(e.Server.Addr, "current server")
		oapi.AddJWTBearerAuth("Authorization")
		oapi.Describe(`
	API Document Description

		function some() {return "code"}
	`)

		e.Add(endpoint.Echo(
			endpoint.Get("/"),
			oapi.Route("greeting", `This route is not authenticated`),
			func(in endpoint.EndpointInput[any, any, ContextQ, any]) (
				res endpoint.DataResponse[endpoint.SingleItemData[string]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "Greeting"
				res.Data.Item = "hello world"
				return res, nil
			},
		))
		gJWT.Add(endpoint.Echo(
			endpoint.Get("/api/collection"),
			oapi.Route("collection.List", `Lists all collections`),
			func(in endpoint.EndpointInput[*Claims, any, struct {
				ContextQ
				FilterCollection
			}, any]) (
				res endpoint.DataResponse[endpoint.CollectionItemData[Collection]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "CollectionList"
				res.Data.Items = []Collection{{Name: "Collection created"}}
				return res, nil
			},
		))

		gJWT.Add(endpoint.Echo(
			endpoint.Get("/api/collection/:id"),
			oapi.Route("collection.Get", `Get one collection`),
			func(in endpoint.EndpointInput[*Claims, struct {
				ID int64 `json:"id,string"`
			}, ContextQ, any]) (
				res endpoint.DataResponse[endpoint.SingleItemData[Collection]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "Collection"
				res.Data.Item = Collection{Name: "Collection"}
				return res, nil
			},
		))
		gJWT.Add(endpoint.Echo(
			endpoint.Put("/api/collection/:id"),
			oapi.Route("collection.Put", `Update a collection`),
			func(in endpoint.EndpointInput[*Claims, struct {
				ID int64 `json:"id,string"`
			}, ContextQ, Collection]) (
				res endpoint.DataResponse[endpoint.SingleItemData[Collection]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "Collection"
				res.Data.Item = Collection{Name: "Collection edited"}
				return res, nil
			},
		),
		)

		gJWT.Add(endpoint.Echo(
			endpoint.Post("/api/collection"),
			oapi.Route("collection.Dreate", `Create a collection`),
			func(in endpoint.EndpointInput[*Claims, any, ContextQ, Collection]) (
				res endpoint.DataResponse[endpoint.SingleItemData[Collection]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "Collection"
				res.Data.Item = Collection{Name: "Collection created"}
				return res, nil
			},
		))

		gJWT.Add(endpoint.Echo(
			endpoint.Delete("/api/collection/:id"),
			oapi.Route("collection.Delete", `Delete one collection by id`),
			func(in endpoint.EndpointInput[*Claims, struct {
				ID int64 `json:"id,string"`
			}, ContextQ, struct{}]) (
				res endpoint.DataResponse[endpoint.SingleItemData[Collection]], err error) {

				res.Context = in.Query.Context
				res.Data.DataDetail.Kind = "Collection"
				res.Data.Item = Collection{Name: "Collection Removed"}
				return res, nil
			},
		))

		swagapijson, err := oapi.T().MarshalJSON()
		if err != nil {
			panic(err)
		}

		e.GET("/docs/swagger.json", func(c echo.Context) error {
			return c.JSON(http.StatusOK, json.RawMessage(swagapijson))
		})
		e.GET("/docs", func(c echo.Context) error {
			return c.HTML(http.StatusOK, `<!DOCTYPE html>
	<html>
	<head>
		<title>API Reference</title>
		<meta charset="utf-8" />
		<meta
		name="viewport"
		content="width=device-width, initial-scale=1" />
		<style>
		body {
			margin: 0;
		}
		</style>
	</head>
	<body>
		<!-- Add your own OpenAPI/Swagger spec file URL here: -->
		<script
		id="api-reference"
		data-url="/docs/swagger.json"></script>
		<script src="https://www.unpkg.com/@scalar/api-reference"></script>
	</body>
	</html>`)
		})
		start := func() error { return e.Start(":8888") }

		go func() {
			err := start()
			if err != nil && err != http.ErrServerClosed {
				e.Logger.Fatalf("server shutdown %s", err)
			} else {
				e.Logger.Fatal("shutting down the server")
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.Shutdown(ctx)
	}
