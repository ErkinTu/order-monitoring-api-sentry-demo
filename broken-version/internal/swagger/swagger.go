package swagger

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const openAPIDocument = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Order Monitoring API - Broken Version",
    "version": "1.0.0",
    "description": "Demo API with intentional bugs for monitoring, Sentry, Elastic and Postman presentation scenarios."
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Broken version"
    }
  ],
  "tags": [
    {
      "name": "System",
      "description": "Service health"
    },
    {
      "name": "Products",
      "description": "Product catalog"
    },
    {
      "name": "Orders",
      "description": "Order operations"
    },
    {
      "name": "Payments",
      "description": "Payment operations"
    }
  ],
  "paths": {
    "/health": {
      "get": {
        "tags": ["System"],
        "summary": "Health check",
        "responses": {
          "200": {
            "description": "API is available",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/HealthResponse"
                }
              }
            }
          }
        }
      }
    },
    "/products": {
      "get": {
        "tags": ["Products"],
        "summary": "List products",
        "responses": {
          "200": {
            "description": "Product list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Product"
                  }
                }
              }
            }
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/orders": {
      "post": {
        "tags": ["Orders"],
        "summary": "Create order",
        "description": "In broken-version, empty order_number is replaced with ORDER-DEMO-DUPLICATE, so repeated requests can fail with a database unique constraint error.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/CreateOrderRequest"
              },
              "examples": {
                "success": {
                  "summary": "Unique order number",
                  "value": {
                    "product_id": 1,
                    "quantity": 1,
                    "customer_email": "student@example.com",
                    "order_number": "BROKEN-DEMO-001"
                  }
                },
                "duplicateBug": {
                  "summary": "Duplicate bug demo",
                  "value": {
                    "product_id": 1,
                    "quantity": 1,
                    "customer_email": "duplicate@example.com"
                  }
                },
                "slowRequest": {
                  "summary": "Slow request demo",
                  "value": {
                    "product_id": 1,
                    "quantity": 1,
                    "customer_email": "slow@example.com",
                    "order_number": "BROKEN-SLOW-001",
                    "simulate_slow": true
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Order created",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Order"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/orders/{id}": {
      "get": {
        "tags": ["Orders"],
        "summary": "Get order by ID",
        "parameters": [
          {
            "$ref": "#/components/parameters/OrderID"
          }
        ],
        "responses": {
          "200": {
            "description": "Order",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Order"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/orders/discount": {
      "post": {
        "tags": ["Orders"],
        "summary": "Calculate discount",
        "description": "In broken-version, discount_percent = 0 causes an integer divide by zero panic.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/DiscountRequest"
              },
              "examples": {
                "success": {
                  "value": {
                    "price": 1000,
                    "discount_percent": 10
                  }
                },
                "panic": {
                  "value": {
                    "price": 1000,
                    "discount_percent": 0
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Discount result",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/DiscountResponse"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    },
    "/orders/{id}/payments": {
      "post": {
        "tags": ["Payments"],
        "summary": "Capture payment",
        "description": "In broken-version, concurrent requests can create multiple captured payments for one order.",
        "parameters": [
          {
            "$ref": "#/components/parameters/OrderID"
          }
        ],
        "responses": {
          "201": {
            "description": "Payment captured",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Payment"
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "409": {
            "$ref": "#/components/responses/Conflict"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      },
      "get": {
        "tags": ["Payments"],
        "summary": "List order payments",
        "parameters": [
          {
            "$ref": "#/components/parameters/OrderID"
          }
        ],
        "responses": {
          "200": {
            "description": "Payment list",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Payment"
                  }
                }
              }
            }
          },
          "400": {
            "$ref": "#/components/responses/BadRequest"
          },
          "404": {
            "$ref": "#/components/responses/NotFound"
          },
          "500": {
            "$ref": "#/components/responses/InternalServerError"
          }
        }
      }
    }
  },
  "components": {
    "parameters": {
      "OrderID": {
        "name": "id",
        "in": "path",
        "required": true,
        "schema": {
          "type": "integer",
          "minimum": 1
        },
        "example": 1
      }
    },
    "responses": {
      "BadRequest": {
        "description": "Invalid request",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "NotFound": {
        "description": "Resource not found",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "Conflict": {
        "description": "Conflict",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      },
      "InternalServerError": {
        "description": "Unexpected server error",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      }
    },
    "schemas": {
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": {
            "type": "string",
            "example": "ok"
          }
        }
      },
      "Product": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "example": 1
          },
          "name": {
            "type": "string",
            "example": "Keyboard"
          },
          "price": {
            "type": "integer",
            "example": 1000
          }
        }
      },
      "CreateOrderRequest": {
        "type": "object",
        "required": ["product_id", "quantity", "customer_email"],
        "properties": {
          "product_id": {
            "type": "integer",
            "example": 1
          },
          "quantity": {
            "type": "integer",
            "example": 1
          },
          "customer_email": {
            "type": "string",
            "example": "student@example.com"
          },
          "order_number": {
            "type": "string",
            "example": "BROKEN-DEMO-001"
          },
          "simulate_slow": {
            "type": "boolean",
            "example": false
          }
        }
      },
      "Order": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "example": 1
          },
          "order_number": {
            "type": "string",
            "example": "BROKEN-DEMO-001"
          },
          "product_id": {
            "type": "integer",
            "example": 1
          },
          "quantity": {
            "type": "integer",
            "example": 1
          },
          "customer_email": {
            "type": "string",
            "example": "student@example.com"
          },
          "total_price": {
            "type": "integer",
            "example": 1000
          },
          "status": {
            "type": "string",
            "example": "created"
          },
          "payment_status": {
            "type": "string",
            "example": "pending"
          },
          "paid_amount": {
            "type": "integer",
            "example": 0
          }
        }
      },
      "DiscountRequest": {
        "type": "object",
        "required": ["price", "discount_percent"],
        "properties": {
          "price": {
            "type": "integer",
            "example": 1000
          },
          "discount_percent": {
            "type": "integer",
            "example": 10
          }
        }
      },
      "DiscountResponse": {
        "type": "object",
        "properties": {
          "price": {
            "type": "integer",
            "example": 1000
          },
          "discount_percent": {
            "type": "integer",
            "example": 10
          },
          "discount_amount": {
            "type": "integer",
            "example": 100
          },
          "final_price": {
            "type": "integer",
            "example": 900
          }
        }
      },
      "Payment": {
        "type": "object",
        "properties": {
          "id": {
            "type": "integer",
            "example": 1
          },
          "order_id": {
            "type": "integer",
            "example": 1
          },
          "amount": {
            "type": "integer",
            "example": 1000
          },
          "status": {
            "type": "string",
            "example": "captured"
          },
          "provider_reference": {
            "type": "string",
            "example": "pay_1710000000000000000"
          }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "error": {
            "type": "string",
            "example": "failed to create order"
          },
          "details": {
            "type": "string",
            "example": "duplicate key value violates unique constraint"
          }
        }
      }
    }
  }
}`

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Order Monitoring API - Broken Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui"
      });
    };
  </script>
</body>
</html>`

func Register(r *gin.Engine) {
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
	})
	r.GET("/swagger/doc.json", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(openAPIDocument))
	})
}
