package gcs

import (
	"fmt"
)

type Drainer struct {
	bucketPairs       map[string]BucketPair
	executionStrategy Strategy
}

func NewDrainer(bucketPairs map[string]BucketPair, executionStrategy Strategy) Restorer {
	return Restorer{
		bucketPairs:       bucketPairs,
		executionStrategy: executionStrategy,
	}
}

func (r Restorer) Drain(backups map[string]BucketBackup) error {
	for bucketIdentifier := range backups {
		_, exists := r.bucketPairs[bucketIdentifier]
		if !exists {
			return fmt.Errorf("bucket identifier '%s' not found in bucketPairs configuration", bucketIdentifier)
		}
	}

	for bucketIdentifier, backup := range backups {
		bucket := r.bucketPairs[bucketIdentifier]

		errs := r.executionStrategy.Run(backup.Blobs, func(blob Blob) error {
			return bucket.backupBucket.CopyVersion(blob, backup.LiveBucketName)
		})

		if len(errs) != 0 {
			return formatErrors(fmt.Sprintf("failed to drain bucket '%s'", bucket.liveBucket.Name()), errs)
		}
	}

	return nil
}
