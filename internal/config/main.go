package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/guregu/null"
	"github.com/jtmilanest/cognito-backup/internal/types"
	log "github.com/sirupsen/logrus"
)

type ConfigParam struct {
	AWSRegion string

	CognitoUserPoolID string
	CognitoRegion     string

	S3BucketName   string
	S3BucketRegion string

	KMSKeyName string
	KMSRegion  string

	BackupPrefix string

	RotationEnabled   null.Bool
	RotationDaysLimit int64
}

// Helper function to verify Environment Variables
func getLookEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Instantiate New Config Parameters
func NewConfigParam(eventRaw interface{}) (*ConfigParam, error) {
	// Instantiate empty ConfigParam
	var config = &ConfigParam{}

	var getFromEvent bool
	var event types.Event

	// Switch between parameter or environment variables
	switch value := eventRaw.(type) {
	case types.Event:
		getFromEvent = true
		event = value
	default:
		getFromEvent = false
	}

	// Process AWS Region
	if awsRegion := getLookEnv("AWS_REGION", ""); awsRegion != "" {
		config.AWSRegion = awsRegion
	} else {
		log.Warn("Environment variable for AWS_REGION is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.AWSRegion != "" {
			config.AWSRegion = event.AWSRegion
		} else {
			log.Warn("Event contains empty awsRegion variable")
		}
	}
	if config.AWSRegion == "" {
		return nil, fmt.Errorf("awsRegion is empty;Configure it via 'AWS_REGION' env variable OR pass in event body")
	}
	// Process AWS Region

	// Process Cognito User Pool ID
	if cognitoUserPoolID := getLookEnv("COGNITO_USER_POOL_ID", ""); cognitoUserPoolID != "" {
		config.CognitoUserPoolID = cognitoUserPoolID
	} else {
		log.Warn("Environment variable for COGNITO_USER_POOL_ID is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.CognitoUserPoolID != "" {
			config.CognitoUserPoolID = event.CognitoUserPoolID
		} else {
			log.Warn("Event contains empty cognitoUserPoolID variable")
		}
	}
	if config.CognitoUserPoolID == "" {
		return nil, fmt.Errorf("cognitoUserPoolID is empty;Configure it via 'COGNITO_USER_POOL_ID' env variable OR pass in event body")
	}
	// Process Cognito User Pool ID

	// Process Cognito Region
	if cognitoRegion := getLookEnv("COGNITO_REGION", ""); cognitoRegion != "" {
		config.CognitoRegion = cognitoRegion
	} else {
		log.Warn("Environment variable for COGNITO_REGION is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.CognitoRegion != "" {
			config.CognitoRegion = event.CognitoRegion
		} else {
			log.Warn("Event contains empty cognitoRegion variable")
		}
	}
	if config.CognitoRegion == "" {
		return nil, fmt.Errorf("cognitoRegion is empty;Configure it via 'COGNITO_REGION' env variable OR pass in event body")
	}
	// Process Cognito Region

	// Process S3BucketName
	if s3BucketName := getLookEnv("S3_BUCKET_NAME", ""); s3BucketName != "" {
		config.S3BucketName = s3BucketName
	} else {
		log.Warn("Environment variable for S3_BUCKET_NAME is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.S3BucketName != "" {
			config.S3BucketName = event.S3BucketName
		} else {
			log.Warn("Event contains empty s3BucketName variable")
		}
	}
	if config.S3BucketName == "" {
		return nil, fmt.Errorf("s3BucketName is empty;Configure it via 'S3_BUCKET_NAME' env variable OR pass in event body")
	}
	// Process S3BucketName

	// Process S3BucketRegion
	if s3BucketRegion := getLookEnv("S3_BUCKET_REGION", ""); s3BucketRegion != "" {
		config.S3BucketRegion = s3BucketRegion
	} else {
		log.Warn("Environment variable for S3_BUCKET_REGION is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.S3BucketRegion != "" {
			config.S3BucketRegion = event.S3BucketRegion
		} else {
			log.Warn("Event contains empty s3BucketRegion variable")
		}
	}
	if config.S3BucketRegion == "" {
		return nil, fmt.Errorf("s3BucketRegion is empty;Configure it via 'S3_BUCKET_REGION' env variable OR pass in event body")
	}
	// Process S3BucketRegion

	// Process BackupPrefix
	if backupPrefix := getLookEnv("BACKUP_PREFIX", ""); backupPrefix != "" {
		config.BackupPrefix = backupPrefix
	} else {
		log.Warn("Environment variable 'BACKUP_PREFIX' is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.BackupPrefix == "" {
			log.Warn("Event contains empty backupPrefix")
		} else {
			config.BackupPrefix = event.BackupPrefix
		}
	}
	// Process BackupPrefix

	// Process RotationEnabled
	if rotationEnabled := getLookEnv("ROTATION_ENABLED", ""); rotationEnabled != "" {

		rotationEnabledValue, err := strconv.ParseBool(rotationEnabled)
		if err != nil {
			return nil, fmt.Errorf("Could not parse 'ROTATION_ENABLED' variable. Error: %w", err)
		}

		config.RotationEnabled = null.NewBool(rotationEnabledValue, true)
	} else {
		log.Warn("Environment variable 'ROTATION_ENABLED' is empty")
	}

	// pass the value to config
	if getFromEvent {
		if event.RotationEnabled.Valid {
			config.RotationEnabled = event.RotationEnabled
		}
	}
	if !config.RotationEnabled.Valid {
		log.Warn("rotationEnabled is not specified, Rotation will be disaled")
		config.RotationEnabled = null.NewBool(false, true)
	}

	// Process RotationDaysLimit
	if config.RotationEnabled.Bool {
		if rotationDaysLimit := getLookEnv("ROTATION_DAYS_LIMIT", ""); rotationDaysLimit != "" {
			rotationDaysValue, err := strconv.ParseInt(rotationDaysLimit, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("Could not parse 'ROTATION_DAYS_LIMIT' variable. Error %w", err)
			}

			config.RotationDaysLimit = rotationDaysValue
		} else {
			log.Warnf("Environment variable 'ROTATION_DAYS_LIMIT' is empty")
		}

		if getFromEvent {
			if event.RotationDaysLimit.Valid {
				config.RotationDaysLimit = event.RotationDaysLimit.Int64
			}
		}

		if config.RotationDaysLimit == 0 {
			return nil, fmt.Errorf("RotationDaysLimit variable should be greater than 0")
		}

	}
	// Process RotationEnabled

	// TODO KMS
	// Process KMSKeyName
	if kmsKeyName := getLookEnv("KMS_KEY_NAME", ""); kmsKeyName != "" {
		config.KMSKeyName = kmsKeyName
	} else {
		log.Warn("Environment variable for 'KMS_KEY_NAME' is empty")
	}

	// pass the value to config from Event/parameter in lambda from Event/parameter in lambda
	if getFromEvent {
		if event.KMSKeyName != "" {
			config.KMSKeyName = event.KMSKeyName
		} else {
			log.Warn("Event contains empty kmsKeyName variable")
		}
	}
	if config.KMSKeyName == "" {
		return nil, fmt.Errorf("kmsKeyName is empty;Configure it via 'KMS_KEY_NAME' env variable OR pass in event body")
	}
	// Process KMSKeyName

	// Process KMSRegion
	if kmsRegion := getLookEnv("KMS_REGION", ""); kmsRegion != "" {
		config.KMSRegion = kmsRegion
	} else {
		log.Warn("Environment variable for KMS_REGION is empty")
	}

	// pass the value to config from Event/parameter in lambda from Event/parameter in lambda
	if getFromEvent {
		if event.KMSRegion != "" {
			config.KMSRegion = event.KMSRegion
		} else {
			log.Warn("Event contains empty kmsRegion variable")
		}
	}
	if config.KMSRegion == "" {
		return nil, fmt.Errorf("kmsRegion is empty;Configure it via 'KMS_REGION' env variable OR pass in event body")
	}
	// Process KMSRegion

	// Return config, no errors
	return config, nil
}
