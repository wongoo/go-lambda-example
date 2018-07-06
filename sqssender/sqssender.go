package main

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/golang/glog"
	"encoding/json"
	"flag"
	"os"
	"github.com/wongoo/go-lambda-example/sqsutil"
)

func SqsSendBatch(sqsSvc *sqs.SQS, entries [] *sqs.SendMessageBatchRequestEntry, queueUrl *string) {
	params := &sqs.SendMessageBatchInput{
		Entries:  entries,
		QueueUrl: queueUrl,
	}

	sendResult, err := sqsSvc.SendMessageBatch(params)

	if err != nil {
		glog.Errorf("Error", err)
	}

	glog.Infof("Send SQS	batch successful : %v", len(sendResult.Successful))
	glog.Infof("Send SQS	batch failed : %v", len(sendResult.Failed))
}

func NewSqsEntry(msg string) *sqs.SendMessageBatchRequestEntry {
	id, _ := uuid.NewV4()
	message := &sqs.SendMessageBatchRequestEntry{
		Id: aws.String(id.String()),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"contentType": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("application/json"),
			},
			"Author": &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String("test"),
			},
		},
		MessageBody: aws.String(string(msg)),
	}
	fmt.Println("new sqs request:", id)
	return message
}

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

	request := &sqsutil.StockRecycleRequest{
		ProductId: "product_1",
		OrderId:   "order_1",
	}
	b, _ := json.Marshal(request)
	msg := string(b)

	msgList := []*sqs.SendMessageBatchRequestEntry{
		NewSqsEntry(msg),
	}
	SqsSendBatch(svc, msgList, qUrl)
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
