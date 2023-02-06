// Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplategeneral = `{
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
        "/request": {
            "get": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Host",
                        "name": "host",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/studyPlaces": {
            "get": {
                "responses": {}
            }
        },
        "/studyPlaces/self": {
            "get": {
                "responses": {}
            }
        },
        "/studyPlaces/{id}": {
            "get": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Study Place ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/uptime": {
            "get": {
                "responses": {}
            }
        }
    }
}`

// SwaggerInfogeneral holds exported Swagger Info so clients can modify it
var SwaggerInfogeneral = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "general",
	SwaggerTemplate:  docTemplategeneral,
}

func init() {
	swag.Register(SwaggerInfogeneral.InstanceName(), SwaggerInfogeneral)
}
