package unversioned

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
)

type Restorer struct {
	bucketPairs map[string]RestoreBucketPair
	artifact    incremental.Artifact
}

func NewRestorer(bucketPairs map[string]RestoreBucketPair, artifact incremental.Artifact) Restorer {
	return Restorer{
		bucketPairs: bucketPairs,
		artifact:    artifact,
	}
}

func (b Restorer) Run() error {
	bucketBackups, err := b.artifact.Load()
	if err != nil {
		return err
	}

	for key := range bucketBackups {
		_, exists := b.bucketPairs[key]
		if !exists {
			return fmt.Errorf(
				"restore config does not mention bucket: %s, but is present in the artifact",
				key,
			)
		}
	}

	for key, pair := range b.bucketPairs {
		bucketBackup, exists := bucketBackups[key]
		if !exists {
			return fmt.Errorf("cannot restore bucket %s, not found in backup artifact", key)
		}

		if len(bucketBackup.Blobs) != 0 {
			err = pair.Restore(bucketBackup)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
