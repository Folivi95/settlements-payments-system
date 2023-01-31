package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateAwsSession(disableSSL bool, awsRegion string, awsEndpoint string) (*session.Session, error) {
	cfgs := &aws.Config{
		Region:           aws.String(awsRegion),
		Endpoint:         aws.String(awsEndpoint),
		S3ForcePathStyle: aws.Bool(true),
		DisableSSL:       aws.Bool(disableSSL),
	}

	sess, err := session.NewSession(cfgs)
	if err != nil {
		return nil, err
	}
	return sess, nil
}
