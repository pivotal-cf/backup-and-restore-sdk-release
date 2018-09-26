package gcs

import (
	"fmt"
	"time"
)

type Backuper struct {
	bucketPairs map[string]BucketPair
}

func NewBackuper(bucketPairs map[string]BucketPair) Backuper {
	return Backuper{
		bucketPairs: bucketPairs,
	}
}

func (b *Backuper) Backup() error {
	for bucketIdentifier, bucketPair := range b.bucketPairs {
		liveBlobs, err := bucketPair.liveBucket.ListBlobs()
		if err != nil {
			return err
		}

		lastBackupBlobs, err := bucketPair.backupBucket.ListLastBackupBlobs()
		if err != nil {
			return err
		}
		backupBlobNames := make(map[string]bool)
		for _, blob := range lastBackupBlobs {
			fmt.Printf("Found backup blob '%+v'\n", blob)
			backupBlobNames[blob.Name] = true
		}

		timestamp := time.Now().Format("2006_01_02_15_04_05")
		newBackupLocation := fmt.Sprintf("%s/%s", timestamp, bucketIdentifier)
		fmt.Printf("Backup location: '%s'\n", newBackupLocation)

		for _, liveBlob := range liveBlobs {
			if !backupBlobNames[liveBlob.Name] {
				fmt.Printf("Copying live blob '%+v' to backup bucket\n", liveBlob)
				err = bucketPair.liveBucket.CopyBlob(bucketPair.backupBucket, newBackupLocation, liveBlob)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
