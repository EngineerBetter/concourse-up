package aws

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// EnsureBucketExists checks if the named bucket exists and creates it if it doesn't
func EnsureBucketExists(name, region string) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}

	client := s3.New(sess, &aws.Config{Region: &region})

	_, err = client.HeadBucket(&s3.HeadBucketInput{Bucket: &name})
	if err == nil {
		return nil
	}

	if err.(awserr.Error).Code() != s3.ErrCodeNoSuchBucket && err.(awserr.Error).Code() != "NotFound" {
		return err
	}

	_, err = client.CreateBucket(&s3.CreateBucketInput{
		Bucket: &name,
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: &region,
		},
	})
	if err != nil {
		return err
	}

	versioningStatus := "Enabled"
	_, err = client.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: &name,
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: &versioningStatus,
		},
	})
	if err != nil {
		return err
	}

	return err
}

// EnsureFileExists checks for the named file in S3 and creates it if it doesn't
func EnsureFileExists(bucket, path, region string, defaultContents []byte) ([]byte, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
	}

	client := s3.New(sess, &aws.Config{Region: &region})

	output, err := client.GetObject(&s3.GetObjectInput{Bucket: &bucket, Key: &path})
	if err == nil {
		contents, err := ioutil.ReadAll(output.Body)
		if err != nil {
			return nil, err
		}

		return contents, nil
	}

	if err.(awserr.Error).Code() != s3.ErrCodeNoSuchKey {
		return nil, err
	}

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   bytes.NewReader(defaultContents),
	})
	if err != nil {
		return nil, err
	}

	return defaultContents, nil
}
