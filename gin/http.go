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

const (
	HeaderCustom = "x-custom"
	HeaderClient = "x-client"
)

var (
	initialized = false
	ginLambda   *ginadapter.GinLambda
	ginEngine   *gin.Engine
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

// Proxy receives an API Gateway proxy event, transforms it into an http.Request
// object, and sends it to the gin.Engine for routing.
// It returns a proxy response object gneerated from the http.ResponseWriter.
func Proxy(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ginRequest, err := ginLambda.ProxyEventToHTTPRequest(req)

	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Could not convert proxy event to request: %v", err)
	}

	addCustomHeader(ginRequest)

	respWriter := core.NewProxyResponseWriter()
	ginEngine.ServeHTTP(http.ResponseWriter(respWriter), ginRequest)

	proxyResponse, err := respWriter.GetProxyResponse()
	if err != nil {
		return core.GatewayTimeout(), core.NewLoggedError("Error while generating proxy response: %v", err)
	}

	return proxyResponse, nil
}

func addCustomHeader(request *http.Request) {
	request.Header.Add(HeaderCustom, "hello")
}

func handleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if !initialized {
		// stdout and stderr are sent to AWS CloudWatch Logs
		log.Printf("Gin cold start\n")

		ginEngine = gin.Default()
		ginEngine.Use(CORSMiddleware())

		ginEngine.GET("/gin", func(context *gin.Context) {
			log.Println("/index")
			log.Println(HeaderCustom, context.GetHeader(HeaderCustom))
			log.Println(HeaderClient, context.Request.Header.Get(HeaderClient))
			context.String(http.StatusOK, "index")
		})

		ginEngine.GET("/gin/env", func(context *gin.Context) {
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
		ginEngine.GET("/gin/hello/:name", func(context *gin.Context) {
			log.Println("/gin/hello")
			name := context.Params.ByName("name")
			context.String(http.StatusOK, "hello %v", name)
		})

		ginLambda = ginadapter.New(ginEngine)
		initialized = true
	}
	// If no name is provided in the HTTP request body, throw an error
	response, err := Proxy(request)

	// Enable CORS
	response.Headers["Access-Control-Allow-Origin"] = "*"
	return response, err
}

func main() {
	lambda.Start(handleRequest)
}
