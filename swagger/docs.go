// Package swagger Code generated by swaggo/swag. DO NOT EDIT
package swagger

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "email": "support@example.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/html"
                ],
                "tags": [
                    "Info"
                ],
                "summary": "Show collected metrics",
                "operationId": "infoMetrics",
                "responses": {
                    "200": {
                        "description": "HTML response with the collected metrics",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/value/{metricType}/{metricName}": {
            "get": {
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Metrics"
                ],
                "summary": "Retrieve a metric's value by type and name",
                "operationId": "getMetricValue",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Metric Type",
                        "name": "metricType",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Metric Name",
                        "name": "metricName",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Metric value returned successfully",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "400": {
                        "description": "Bad Request - Unknown metric type",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "404": {
                        "description": "Not Found - Metric not found",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "tags": [
        {
            "description": "\"Endpoints for retrieving the status and information of the service.\"",
            "name": "Info"
        },
        {
            "description": "\"Endpoints for managing and accessing metric data stored in the service.\"",
            "name": "Metric Storage"
        }
    ]
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Metrics API",
	Description:      "Metrics service for monitoring, retrieving, and managing metric data.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
