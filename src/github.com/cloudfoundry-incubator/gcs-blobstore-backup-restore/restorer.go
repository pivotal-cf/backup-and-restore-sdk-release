package gcs

type Restorer struct {
	buckets           map[string]BucketPair
	executionStrategy Strategy
}

func NewRestorer(buckets map[string]BucketPair, executionStrategy Strategy) Restorer {
	return Restorer{
		buckets:           buckets,
		executionStrategy: executionStrategy,
	}
}

func (r Restorer) Restore(backups map[string]BucketBackup) error {
	return nil
}
