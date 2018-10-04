package main

import (
	"errors"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	serviceAccount := os.Getenv("GCP_SERVICE_ACCOUNT_KEY")
	if serviceAccount == "" {
		log.Fatalln("must set GCP_SERVICE_ACCOUNT_KEY")
	}

	liveBucket, err := gcs.NewSDKBucket(serviceAccount, "gcs-spike-live-bucket", "droplets")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("listing live blobs...")
	liveBlobs, err := liveBucket.ListBlobs()

	backupBucket, err := gcs.NewSDKBucket(serviceAccount, "gcs-spike-backup-bucket", "droplets")
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("finding last path path...")
	prefix, err := backupBucket.LastBackupPrefix()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("listing last backup path blobs...")
	backupBlobs, err := liveBucket.ListBlobs()

	log.Println("creating delta of blobs...")
	currentBlobNames := make(map[string]bool)
	for _, blob := range backupBlobs {
		currentBlobNames[blob.Name] = true
	}

	executionStrategy := gcs.NewParallelStrategy()

	log.Println("starting to copy delta blobs...")
	errs := executionStrategy.Run(liveBlobs, func(blob gcs.Blob) error {
		if currentBlobNames[blob.Name] {
			return nil
		}
		return errors.New("should not be here, live blobs != backup blobs")
	})

	if len(errs) != 0 {
		log.Fatalf("failed to copy blobs to backup location '%s': %+v", prefix, errs)
	}

	log.Printf("prefix: '%s'", prefix)
}
