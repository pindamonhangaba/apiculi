package main

import (
	"log"

	"github.com/pindamonhangaba/apiculi/endpoint"

	"github.com/gofiber/fiber/v2"
)

type endpointResponse struct {
	CreatedResourceName string `json:"created_resource_name"`
}
type endpointInput = endpoint.EndpointInput[struct {
	Name string `json:"name"`
}, struct {
	ID string `json:"id"`
}, struct {
	Context string `json:"context"`
	Number  string `json:"number"`
}, struct {
	ResourceName string `json:"resource_name"`
}]

func main() {
	listen := "localhost:3000"
	oapi := endpoint.NewOpenAPI("Endpoint Docs", "v1.0.1")
	oapi.AddServer(listen, "current server")

	app := fiber.New()

	app.Add(endpoint.Fiber(
		endpoint.Post("/api/endpoint/{id}"),
		oapi.Describe("Create one resource", `
			* Nice list
			* Description ~goes here~
		`),
		func(in endpointInput) endpoint.DataResponse[endpoint.SingleItemData[endpointResponse]] {
			return endpoint.DataResponse[endpoint.SingleItemData[endpointResponse]]{
				Context: in.Query.Context,
				Data: endpoint.SingleItemData[endpointResponse]{
					DataDetail: endpoint.DataDetail{
						Kind: "resource",
					},
					Item: endpointResponse{
						CreatedResourceName: in.Body.ResourceName,
					},
				},
			}
		},
	))

	swagapijson, err := oapi.T().MarshalJSON()
	if err != nil {
		panic(err)
	}

	app.Get("/swagger.json", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		return c.Send(swagapijson)
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(`
			<!DOCTYPE html>
			<html>
			<head>
				<title>Redoc</title>
				<!-- needed for adaptive design -->
				<meta charset="utf-8"/>
				<meta name="viewport" content="width=device-width, initial-scale=1">
				<link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

				<!--
				Redoc doesn't change outer page styles
				-->
				<style>
				body {
					margin: 0;
					padding: 0;
				}
				</style>
			</head>
			<body>
				<redoc spec-url='http://` + listen + `/swagger.json'></redoc>
				<script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"> </script>
			</body>
			</html>
		`)
	})

	log.Fatal(app.Listen(listen))
}
