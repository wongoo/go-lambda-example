package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"flag"
	"os"
	"github.com/wongoo/go-lambda-example/sqsutil"
	"encoding/json"
)

var (
	qName           = flag.String("qname", "", "queue name")
	accessKeyId     = flag.String("key", "", "access key id")
	accessKeySecret = flag.String("secret", "", "access key secret")
)

func main() {
	flag.Parse()                    // To meet glog's requirement.
	flag.Set("logtostderr", "true") // Log to stderr only, instead of file.
	if *qName == "" {
		exitErrorf("qname needed")
	}
	if *accessKeyId == "" {
		exitErrorf("key needed")
	}
	if *accessKeySecret == "" {
		exitErrorf("secret needed")
	}

	svc := sqsutil.SqsConnect(*accessKeyId, *accessKeySecret)
	qUrl, err := sqsutil.SqsQueueUrl(svc, *qName)
	if err != nil {
		exitErrorf(err.Error())
	}
	fmt.Println("queue url:", *qUrl)

	result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl: qUrl,
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MaxNumberOfMessages: aws.Int64(1),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds: aws.Int64(20),
	})
	if err != nil {
		exitErrorf("Unable to receive message from queue %q, %v.", *qUrl, err)
	}

	fmt.Printf("Received %d messages.\n", len(result.Messages))
	if len(result.Messages) > 0 {
		for _, message := range result.Messages {
			handleMessage(svc, qUrl, message)
		}
	}
}

func handleMessage(svc *sqs.SQS, qUrl *string, message *sqs.Message) {
	fmt.Println(message)

	b := []byte(*message.Body)
	srr := &sqsutil.StockRecycleRequest{}
	err := json.Unmarshal(b, srr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("srr:", srr)

	resultDelete, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      qUrl,
		ReceiptHandle: message.ReceiptHandle,
	})

	if err != nil {
		fmt.Println("Delete Error", err)
		return
	}

	fmt.Println("Message Deleted", resultDelete)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
