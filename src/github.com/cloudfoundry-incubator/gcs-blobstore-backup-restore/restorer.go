package gcs

type Restorer struct {
	buckets map[string]BucketPair
}

func NewRestorer(buckets map[string]BucketPair) Restorer {
	return Restorer{
		buckets: buckets,
	}
}

func (r Restorer) Restore(backupArtifact map[string]BackupBucketDirectory) error {
	return nil
}
