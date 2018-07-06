// authors: wangoo
// created: 2018-07-06
// sqs util

package sqsutil

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"errors"
)

func SqsConnect(key, secret string) *sqs.SQS {
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(endpoints.CnNorth1RegionID),
		Credentials: credentials.NewStaticCredentials(key, secret, "")},
	)
	_, err := sess.Config.Credentials.Get()
	if err != nil {
		fmt.Println("config Credentials err:", err.Error())
		return nil
	}
	return sqs.New(sess)
}

func SqsQueueUrl(svc *sqs.SQS, qName string) (qUrl *string, err error) {
	resultURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(qName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			err = errors.New(fmt.Sprintf("Unable to find queue %q.", qName))
			return
		}
		err = errors.New(fmt.Sprintf("Unable to queue %q, %v.", qName, err))
		return
	}
	qUrl = resultURL.QueueUrl
	return

}

type StockRecycleRequest struct {
	ProductId string `json:"product_id"`
	OrderId   string `json:"order_id"`
}
