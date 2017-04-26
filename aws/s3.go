package aws

import (
	"bytes"
	"io/ioutil"

	"time"

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
	awsErrCodeNoSuchBucket = "NoSuchBucket"

	// ErrCodeNoSuchKey for service response error code
	// "NoSuchKey".
	//
	// The specified key does not exist.
	awsErrCodeNoSuchKey = "NoSuchKey"

	// Returned when calling HEAD on non-existant bucket or object
	awsErrCodeNotFound = "NotFound"
)

// DeleteVersionedBucket deletes and empties a versioned bucket
func DeleteVersionedBucket(name, region string) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}

	client := s3.New(sess, &aws.Config{Region: &region})

	// Delete all objects
	objects := []*s3.Object{}
	err = client.ListObjectsPages(&s3.ListObjectsInput{Bucket: &name},
		func(output *s3.ListObjectsOutput, _ bool) bool {
			objects = append(objects, output.Contents...)

			return true
		})
	if err != nil {
		return err
	}

	for _, object := range objects {
		_, err = client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &name,
			Key:    object.Key,
		})
		if err != nil {
			return nil
		}
	}

	time.Sleep(time.Second)

	_, err = client.DeleteBucket(&s3.DeleteBucketInput{Bucket: &name})
	if err != nil {
		return err
	}

	return nil
}

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

	awsErrCode := err.(awserr.Error).Code()
	if awsErrCode != awsErrCodeNotFound && awsErrCode != awsErrCodeNoSuchBucket {
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
		awsErrCode := err.(awserr.Error).Code()
		if awsErrCode == awsErrCodeNotFound || awsErrCode == awsErrCodeNoSuchKey {
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

	awsErrCode := err.(awserr.Error).Code()
	if awsErrCode != awsErrCodeNoSuchKey && awsErrCode != awsErrCodeNotFound {
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

// DeleteFile deletes a file from S3
func DeleteFile(bucket, path, region string) error {
	sess, err := session.NewSession(aws.NewConfig().WithCredentialsChainVerboseErrors(true))
	if err != nil {
		return err
	}

	client := s3.New(sess, &aws.Config{Region: &region})
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &path,
	})

	return err
}
