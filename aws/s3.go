package aws

import (
	"bytes"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// New versions of github.com/aws/aws-sdk-go/aws have these consts
	// but the version currently pinned by bosh-cli v2 does not

	// ErrCodeNoSuchBucket for service response error code
	// "NoSuchBucket".
	//
	// The specified bucket does not exist.
	ErrCodeNoSuchBucket = "NoSuchBucket"

	// ErrCodeNoSuchKey for service response error code
	// "NoSuchKey".
	//
	// The specified key does not exist.
	ErrCodeNoSuchKey = "NoSuchKey"
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

	if err.(awserr.Error).Code() != ErrCodeNoSuchBucket && err.(awserr.Error).Code() != "NotFound" {
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

	return err
}

// WriteFile writes the specified S3 object
func WriteFile(bucket, path, region string, contents []byte) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}
	client := s3.New(sess, &aws.Config{Region: &region})

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   bytes.NewReader(contents),
	})
	return err
}

// HasFile returns true if the specified S3 object exists
func HasFile(bucket, path, region string) (bool, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return false, err
	}
	client := s3.New(sess, &aws.Config{Region: &region})

	_, err = client.HeadObject(&s3.HeadObjectInput{Bucket: &bucket, Key: &path})
	if err != nil {
		errCode := err.(awserr.Error).Code()
		if errCode == ErrCodeNoSuchKey || errCode == "NotFound" {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// EnsureFileExists checks for the named file in S3 and creates it if it doesn't
// Second argument is true if new file was created
func EnsureFileExists(bucket, path, region string, defaultContents []byte) ([]byte, bool, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, false, err
	}

	client := s3.New(sess, &aws.Config{Region: &region})

	output, err := client.GetObject(&s3.GetObjectInput{Bucket: &bucket, Key: &path})
	if err == nil {
		var contents []byte
		contents, err = ioutil.ReadAll(output.Body)
		if err != nil {
			return nil, false, err
		}

		// Successfully loaded file
		return contents, true, nil
	}

	if err.(awserr.Error).Code() != ErrCodeNoSuchKey {
		return nil, false, err
	}

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &path,
		Body:   bytes.NewReader(defaultContents),
	})
	if err != nil {
		return nil, false, err
	}

	// Created file from given contents
	return defaultContents, true, nil
}

// LoadFile loads a file from S3
func LoadFile(bucket, path, region string) ([]byte, error) {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return nil, err
	}

	client := s3.New(sess, &aws.Config{Region: &region})

	output, err := client.GetObject(&s3.GetObjectInput{Bucket: &bucket, Key: &path})
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(output.Body)
}
