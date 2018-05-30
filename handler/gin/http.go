package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"log"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"github.com/awslabs/aws-lambda-go-api-proxy/core"
	"fmt"
)

var (
	initialized = false
	ginLambda   *ginadapter.GinLambda
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !initialized {
		// stdout and stderr are sent to AWS CloudWatch Logs
		log.Printf("Gin cold start\n")

		r := gin.Default()
		r.Use(CORSMiddleware())

		r.GET("/gin", func(context *gin.Context) {
			log.Println("/index")
			context.String(http.StatusOK, "index")
		})

		r.GET("/gin/env", func(context *gin.Context) {
			log.Println("/gin/env")
			ctxBody := " no gw context"
			acc := new(core.RequestAccessor)
			ctx, err := acc.GetAPIGatewayContext(context.Request)
			if err != nil {
				ctxBody += ",Error:" + err.Error()
			} else {
				ctxBody = fmt.Sprint(ctx)
			}

			context.String(http.StatusOK, "current env is %v <br> %v", GIN_GO_ENV, ctxBody)
		})
		r.GET("/gin/hello/:name", func(context *gin.Context) {
			log.Println("/gin/hello")
			name := context.Params.ByName("name")
			context.String(http.StatusOK, "hello %v", name)
		})

		ginLambda = ginadapter.New(r)
		initialized = true
	}
	// If no name is provided in the HTTP request body, throw an error
	response, err := ginLambda.Proxy(request)

	// Enable CORS
	response.Headers["Access-Control-Allow-Origin"] = "*"
	return response, err
}

func main() {
	lambda.Start(handleRequest)
}
