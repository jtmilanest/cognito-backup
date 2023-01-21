package lambda

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/jtmilanest/cognito-backup/internal/cloud"
	"github.com/jtmilanest/cognito-backup/internal/config"
	log "github.com/sirupsen/logrus"
)

// Encrypt S3 Data via KMS
func encryptViaKMS(ctx context.Context, client *cloud.Client, kmsKeyName string, data []byte) (*kms.EncryptOutput, error) {
	result, err := client.KMSClient.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     aws.String(kmsKeyName),
		Plaintext: data,
	})

	return result, err
}

// Backup Data to S3
func uploadToS3(ctx context.Context, client *cloud.Client, bucketName, keyName string, data *kms.EncryptOutput) error {
	_, err := client.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(bucketName),
		Key:                  aws.String(keyName),
		ACL:                  types.ObjectCannedACLPrivate,
		Body:                 bytes.NewReader(data.CiphertextBlob),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: types.ServerSideEncryptionAes256,
		Tagging:              aws.String(fmt.Sprintf("producer=%v", lambdaAwsTag)),
	})

	return err
}

// Helper function for prefixing backup path
func getKeyName(prefix, timestamp, name string) string {
	var output string
	if prefix == "" {
		output = fmt.Sprintf("%v/%v", timestamp, name)
	} else {
		output = fmt.Sprintf("%v/%v/%v", prefix, timestamp, name)
	}
	return output
}

// Helper function for rotating backups
func rotateBackups(ctx context.Context, client *cloud.Client, bucketName string, rotationDaysLimit int64) error {

	now := time.Now()

	objects, err := client.S3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return err
	}

	for _, obj := range objects.Contents {
		diffDays := int64(now.Sub(*obj.LastModified).Hours() / 24)
		if diffDays >= rotationDaysLimit {
			log.Infof("Object %v is %v days old. It's greater (or equal) than rotation days limit (%v days), Due to that it will be deleted", *obj.Key, diffDays, rotationDaysLimit)
			_, err = client.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(bucketName),
				Key:    obj.Key,
			})
			if err != nil {
				log.Warnf("Failed to delete %v from bucket %v. Error: %v", *obj.Key, bucketName, err)
			} else {
				log.Debugf("Object %v has been successfully deleted from bucket %v", *obj.Key, bucketName)
			}
		} else {
			log.Debugf("Object %v is %v days old. It's less than rotation days limit (%v days). Due to that it will be skipped", *obj.Key, diffDays, rotationDaysLimit)
		}
	}

	return nil
}

// Execute Lambda Function
func Execute(ctx context.Context, config config.ConfigParam) error {
	client, err := cloud.New(ctx, config.CognitoRegion, config.S3BucketRegion, config.KMSRegion)
	if err != nil {
		return fmt.Errorf("Could not create AWS client. Error: %w", err)
	}

	timestamp := time.Now().Format(time.RFC3339)

	// Backup Cognito users
	users, err := client.CognitoClient.ListUsers(ctx, &cognitoidentityprovider.ListUsersInput{
		UserPoolId: aws.String(config.CognitoUserPoolID),
	})
	if err != nil {
		return fmt.Errorf("Failed to get list of cognito users. Error: %w", err)
	}

	usersData, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("Failed to marshal cognito users structure. Error: %w", err)
	}

	encUsersData, err := encryptViaKMS(ctx, client, config.KMSKeyName, usersData)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return fmt.Errorf("Failed to encrypt cognito users. Error: %w", err)
	}

	err = uploadToS3(ctx, client, config.S3BucketName, getKeyName(config.BackupPrefix, timestamp, "users.json"), encUsersData)
	if err != nil {
		return fmt.Errorf("Failed to upload cognito users backup to S3. Error: %w", err)
	}

	// Backup Cognito groups
	groups, err := client.CognitoClient.ListGroups(ctx, &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: aws.String(config.CognitoUserPoolID),
	})
	if err != nil {
		return fmt.Errorf("Failed to get list of cognito groups. Error: %w", err)
	}

	groupsData, err := json.Marshal(groups)
	if err != nil {
		return fmt.Errorf("Failed to marshal cognito groups structure. Error: %w", err)
	}

	encGroupData, err := encryptViaKMS(ctx, client, config.KMSKeyName, groupsData)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return fmt.Errorf("Failed to encrypt cognito users. Error: %w", err)
	}

	err = uploadToS3(ctx, client, config.S3BucketName, getKeyName(config.BackupPrefix, timestamp, "groups.json"), encGroupData)
	if err != nil {
		return fmt.Errorf("Failed to upload cognito groups backup to S3. Error: %w", err)
	}

	if config.RotationEnabled.Bool {
		err = rotateBackups(ctx, client, config.S3BucketName, config.RotationDaysLimit)
		if err != nil {

			return fmt.Errorf("Rotation has been failed. Error: %w", err)
		}
	} else {
		log.Warnf("Pay attention that rotation is disabled; If you want to disable it, activate rotation via env variables, or event body")
	}

	return nil
}
