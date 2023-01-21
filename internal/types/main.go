package types

import "github.com/guregu/null"

// Response struct
type Response struct {
	Message string `json:"answer"`
}

// Event structw
type Event struct {
	AWSRegion string `json:"awsRegion"`

	CognitoUserPoolID string `json:"cognitoUserPoolID"`
	CognitoRegion     string `json:"cognitoRegion"`

	S3BucketName   string `json:"s3BucketName"`
	S3BucketRegion string `json:"s3BucketRegion"`

	KMSKeyName string `json:"kmsKeyName"`
	KMSRegion  string `json:"kmsRegion"`

	BackupPrefix string `json:"backupPrefix"`

	RotationEnabled   null.Bool `json:"rotationEnabled"`
	RotationDaysLimit null.Int  `json:"rotationDaysLimit"`
}
