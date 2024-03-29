// Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplateuser = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "responses": {}
            },
            "put": {
                "responses": {}
            }
        },
        "/accept": {
            "get": {
                "responses": {}
            },
            "post": {
                "responses": {}
            }
        },
        "/block": {
            "post": {
                "responses": {}
            }
        },
        "/firebase/token": {
            "put": {
                "responses": {}
            }
        },
        "/password/reset": {
            "put": {
                "responses": {}
            },
            "post": {
                "responses": {}
            }
        }
    }
}`

// SwaggerInfouser holds exported Swagger Info so clients can modify it
var SwaggerInfouser = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "/api/schedule",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "user",
	SwaggerTemplate:  docTemplateuser,
}

func init() {
	swag.Register(SwaggerInfouser.InstanceName(), SwaggerInfouser)
}
