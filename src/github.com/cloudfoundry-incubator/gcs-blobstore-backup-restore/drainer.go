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

func (r Restorer) Drain() error {
	for bucketIdentifier := range r.bucketPairs {
		_, exists := r.bucketPairs[bucketIdentifier]
		if !exists {
			return fmt.Errorf("bucket identifier '%s' not found in bucket configuration", bucketIdentifier)
		}
	}

	for bucketIdentifier := range r.bucketPairs {
		bucketPair := r.bucketPairs[bucketIdentifier]

		// list blobs in previous backup location
		previousBackupBlobs, err := bucketPair.backupBucket.ListPenultimateBackupBlobs()
		if err != nil {
			return err
		}
		// list blobs in current backup location
		currentBackupBlobs, err := bucketPair.backupBucket.ListLastBackupBlobs()
		if err != nil {
			return err
		}
		// copy blobs in previous that are not in current
		currentBlobNames := make(map[string]bool)
		for _, blob := range currentBackupBlobs {
			currentBlobNames[blob.Name] = true
		}

		currentBackupPrefix, err := bucketPair.backupBucket.LastBackupPrefix()
		if err != nil {
			return err
		}

		errs := r.executionStrategy.Run(previousBackupBlobs, func(blob Blob) error {
			if currentBlobNames[blob.Name] {
				return nil
			}
			fmt.Printf("Copying reused backup blob '%+v' to new backup bucket\n", blob)
			return bucketPair.backupBucket.CopyBlob(bucketPair.backupBucket, currentBackupPrefix, blob)
		})

		if len(errs) != 0 {
			return formatErrors(fmt.Sprintf("failed to copy blobs to backup location '%s'", currentBackupPrefix), errs)
		}
	}

	return nil
}
