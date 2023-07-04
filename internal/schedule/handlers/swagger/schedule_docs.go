// Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplateschedule = `{
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
            },
            "post": {
                "responses": {}
            },
            "delete": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Lesson ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/between/{startDate}/{endDate}": {
            "delete": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "From date",
                        "name": "startDate",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "To date",
                        "name": "endDate",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/general": {
            "get": {
                "responses": {}
            }
        },
        "/general/list": {
            "post": {
                "responses": {}
            }
        },
        "/general/{type}/{name}": {
            "get": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Type",
                        "name": "type",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "RoleName",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/getTypes": {
            "get": {
                "responses": {}
            }
        },
        "/lessons/{id}": {
            "get": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Lesson ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/list": {
            "post": {
                "responses": {}
            }
        },
        "/makeCurrent/:date": {
            "post": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Date",
                        "name": "date",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        },
        "/makeGeneral": {
            "post": {
                "responses": {}
            }
        },
        "/{type}/{name}": {
            "get": {
                "parameters": [
                    {
                        "type": "string",
                        "description": "Type",
                        "name": "type",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "RoleName",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {}
            }
        }
    }
}`

// SwaggerInfoschedule holds exported Swagger Info so clients can modify it
var SwaggerInfoschedule = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "/api/schedule",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "schedule",
	SwaggerTemplate:  docTemplateschedule,
}

func init() {
	swag.Register(SwaggerInfoschedule.InstanceName(), SwaggerInfoschedule)
}
