// authors: wangoo
// created: 2018-07-06
// see: https://github.com/aws/aws-lambda-go/blob/master/events/README_SQS.md

package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		fmt.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)
	}

	return nil
}
