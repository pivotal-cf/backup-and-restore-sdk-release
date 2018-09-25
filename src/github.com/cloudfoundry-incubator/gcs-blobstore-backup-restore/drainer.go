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
			return fmt.Errorf("bucket identifier '%s' not found in bucket configuration", bucketIdentifier)
		}
	}

	for bucketIdentifier, backup := range backups {
		bucketPair := r.bucketPairs[bucketIdentifier]

		blobs, err := bucketPair.backupBucket.ListBlobs()
		if err != nil {
			return err
		}

		blobNames := make(map[string]bool)
		for _, blob := range blobs {
			blobNames[blob.Name] = true
		}

		errs := r.executionStrategy.Run(backup.Blobs, func(blob Blob) error {
			if blobNames[blob.Name] {
				return nil
			}
			return bucketPair.backupBucket.CopyVersion(blob, backup.LiveBucketName)
		})

		if len(errs) != 0 {
			return formatErrors(fmt.Sprintf("failed to drain bucket '%s'", bucketPair.liveBucket.Name()), errs)
		}
	}

	return nil
}
