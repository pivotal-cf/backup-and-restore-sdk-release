package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	artifactPath := flag.String("artifact-file", "", "Path to the artifact file")
	configPath := flag.String("config", "", "Path to JSON config file")
	backupAction := flag.Bool("backup", false, "Run blobstore backup")
	drainAction := flag.Bool("drain", false, "Run blobstore drain")
	restoreAction := flag.Bool("restore", false, "Run blobstore restore")

	flag.Parse()

	if !(*backupAction || *drainAction || *restoreAction) {
		log.Fatal("missing --backup, --drain, or --restore flag")
	}

	if (*backupAction && *drainAction) || (*backupAction && *restoreAction) || (*drainAction && *restoreAction) {
		log.Fatal("only one of: --backup, --drain, or --restore can be provided")
	}

	config, err := gcs.ParseConfig(*configPath)
	exitOnError(err)

	bucketPairs, err := gcs.BuildBuckets(config)
	exitOnError(err)

	artifact := gcs.NewArtifact(*artifactPath)

	executionStrategy := gcs.NewParallelStrategy()

	if *backupAction {
		backuper := gcs.NewBackuper(bucketPairs)

		err := backuper.Backup()
		exitOnError(err)
	} else if *drainAction {
		drainer := gcs.NewDrainer(bucketPairs, executionStrategy)

		err = drainer.Drain()
		exitOnError(err)
	} else {
		restorer := gcs.NewRestorer(bucketPairs, executionStrategy)

		backups, err := artifact.Read()
		exitOnError(err)

		err = restorer.Restore(backups)
		exitOnError(err)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
