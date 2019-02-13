package gcs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type BucketPair struct {
	LiveBucket        Bucket
	BackupBucket      Bucket
	ExecutionStrategy executor.Executor
	IDs               []string
}

func BuildBucketPairs(gcpServiceAccountKey string, config map[string]Config) (map[string]BucketPair, error) {
	buckets := map[string]BucketPair{}
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	pairs := map[string]string{}
	for bucketID, bucketConfig := range config {
		bucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BucketName)
		if err != nil {
			return nil, err
		}

		backupBucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BackupBucketName)
		if err != nil {
			return nil, err
		}

		bucketNamePair := bucket.Name() + ":" + backupBucket.Name()
		if dupedBucketId, ok := pairs[bucketNamePair]; ok {
			bucketPair := buckets[dupedBucketId]
			bucketPair.IDs = append(bucketPair.IDs, bucketID)

			delete(buckets, dupedBucketId)

			sort.Strings(bucketPair.IDs)
			dupedBucketId = strings.Join(bucketPair.IDs, "-")

			buckets[dupedBucketId] = bucketPair
			pairs[bucketNamePair] = dupedBucketId
			continue
		}

		buckets[bucketID] = BucketPair{
			LiveBucket:        bucket,
			BackupBucket:      backupBucket,
			ExecutionStrategy: exe,
			IDs:               []string{bucketID},
		}
		fmt.Println(buckets)
		pairs[bucketNamePair] = bucketID
	}

	return buckets, nil
}
