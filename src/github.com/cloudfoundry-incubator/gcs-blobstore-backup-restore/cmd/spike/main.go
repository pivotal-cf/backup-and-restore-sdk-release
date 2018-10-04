package main

import (
	"log"
	"os"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	serviceAccount := os.Getenv("GCP_SERVICE_ACCOUNT_KEY")
	if serviceAccount == "" {
		log.Fatalln("must set GCP_SERVICE_ACCOUNT_KEY")
	}

	bucket, err := gcs.NewSDKBucket(serviceAccount, "gcs-spike-backup-bucket", "droplets")
	if err != nil {
		log.Fatalln(err)
	}

	prefix, err := bucket.LastBackupPrefix()
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("prefix: '%s'", prefix)
}
