package gcs_test

import (
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var firstBucket *fakes.FakeBucket
	var secondBucket *fakes.FakeBucket
	var firstBackupBucket *fakes.FakeBucket
	var secondBackupBucket *fakes.FakeBucket
	var backupArtifact map[string]gcs.BackupBucketDirectory

	var restorer gcs.Restorer

	const firstBucketName = "first-bucket-name"
	const secondBucketName = "second-bucket-name"
	const firstBackupBucketName = "first-bucket-name"
	const secondBackupBucketName = "second-bucket-name"
	const firstBucketID = "droplets"
	const secondBucketID = "buildpacks"
	const timestamp = "2001_01_01_00_00_00"

	BeforeEach(func() {
		firstBucket = new(fakes.FakeBucket)
		secondBucket = new(fakes.FakeBucket)
		firstBackupBucket = new(fakes.FakeBucket)
		secondBackupBucket = new(fakes.FakeBucket)

		firstBucket.NameReturns(firstBucketName)
		secondBucket.NameReturns(secondBucketName)
		firstBackupBucket.NameReturns(firstBackupBucketName)
		secondBackupBucket.NameReturns(secondBackupBucketName)

		backupArtifact = map[string]gcs.BackupBucketDirectory{
			firstBucketID: {
				BucketName: firstBackupBucketName,
				Path:       timestamp + "/" + firstBucketID,
			},
			secondBucketID: {
				BucketName: secondBackupBucketName,
				Path:       timestamp + "/" + secondBucketID,
			},
		}

		restorer = gcs.NewRestorer(map[string]gcs.BucketPair{
			"first":  {Bucket: firstBucket, BackupBucket: firstBackupBucket},
			"second": {Bucket: secondBucket, BackupBucket: secondBackupBucket},
		})
	})
})
