package gcs

type Backuper struct {
	bucketPairs map[string]BucketPair
}

func NewBackuper(bucketPairs map[string]BucketPair) Backuper {
	return Backuper{
		bucketPairs: bucketPairs,
	}
}

func (b *Backuper) Backup() (map[string]BucketBackup, error) {
	bucketBackups := map[string]BucketBackup{}

	for bucketIdentifier, bucketPair := range b.bucketPairs {
		blobs, err := bucketPair.liveBucket.ListBlobs()
		if err != nil {
			return nil, err
		}

		bucketBackups[bucketIdentifier] = BucketBackup{
			LiveBucketName:   bucketPair.liveBucket.Name(),
			BackupBucketName: bucketPair.backupBucket.Name(),
			Blobs:            blobs,
		}
	}

	return bucketBackups, nil
}
