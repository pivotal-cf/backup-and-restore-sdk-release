// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package s3_test

import (
	"time"

	"strconv"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("S3 versioned backup and restore", func() {
	var region string
	var cloneRegion string
	var unversionedRegion string
	var bucket string
	var cloneBucket string
	var unversionedBucket string
	var fileName1 string
	var fileName2 string
	var fileName3 string
	var artifactDirPath string

	var backuperInstance JobInstance
	var backuperInstanceWithClonedBucket JobInstance
	var backuperInstanceWithUnversionedBucket JobInstance
	var backuperInstanceWithCustomCaCertBlobstore JobInstance

	BeforeEach(func() {
		backuperInstance = JobInstance{
			Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
			Name:       "backuper",
			Index:      "0",
		}

		region = MustHaveEnv("AWS_TEST_BUCKET_REGION")
		bucket = MustHaveEnv("AWS_TEST_BUCKET_NAME")

		DeleteAllVersionsFromBucket(region, bucket)

		artifactDirPath = "/tmp/s3-versioned-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		backuperInstance.RunSuccessfully("mkdir -p " + artifactDirPath)
	})

	AfterEach(func() {
		DeleteAllVersionsFromBucket(region, bucket)
		backuperInstance.RunSuccessfully("rm -rf " + artifactDirPath)
	})

	Context("backs up and restores in-place", func() {
		It("succeeds", func() {
			fileName1 = UploadTimestampedFileToBucket(region, bucket, "file1", "FILE1")
			fileName2 = UploadTimestampedFileToBucket(region, bucket, "file2", "FILE2")

			backuperInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/backup")

			DeleteFileFromBucket(region, bucket, fileName1)
			WriteFileInBucket(region, bucket, fileName2, "FILE2_NEW")
			fileName3 = UploadTimestampedFileToBucket(region, bucket, "file3", "FILE3")

			backuperInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/restore")

			filesList := ListFilesFromBucket(region, bucket)
			Expect(filesList).To(ConsistOf(fileName1, fileName2, fileName3))

			Expect(GetFileContentsFromBucket(region, bucket, fileName1)).To(Equal("FILE1"))
			Expect(GetFileContentsFromBucket(region, bucket, fileName2)).To(Equal("FILE2"))
			Expect(GetFileContentsFromBucket(region, bucket, fileName3)).To(Equal("FILE3"))
		})
	})

	Context("backs up and restores to a different bucket", func() {
		BeforeEach(func() {
			backuperInstanceWithClonedBucket = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "clone-backuper",
				Index:      "0",
			}

			cloneRegion = MustHaveEnv("AWS_TEST_CLONE_BUCKET_REGION")
			cloneBucket = MustHaveEnv("AWS_TEST_CLONE_BUCKET_NAME")

			DeleteAllVersionsFromBucket(cloneRegion, cloneBucket)
			backuperInstanceWithClonedBucket.RunSuccessfully("mkdir -p " + artifactDirPath)
		})

		AfterEach(func() {
			DeleteAllVersionsFromBucket(cloneRegion, cloneBucket)
			backuperInstanceWithClonedBucket.RunSuccessfully("rm -rf " + artifactDirPath)
		})

		It("succeeds", func() {
			fileName1 = UploadTimestampedFileToBucket(region, bucket, "file1", "FILE1")
			fileName2 = UploadTimestampedFileToBucket(region, bucket, "file2", "FILE2")

			backuperInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/backup")

			backuperInstance.Download(artifactDirPath+"/blobstore.json", "/tmp/blobstore.json")
			backuperInstanceWithClonedBucket.Upload("/tmp/blobstore.json", artifactDirPath+"/blobstore.json")

			backuperInstanceWithClonedBucket.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/restore")

			filesList := ListFilesFromBucket(cloneRegion, cloneBucket)
			Expect(filesList).To(ConsistOf(fileName1, fileName2))

			Expect(GetFileContentsFromBucket(cloneRegion, cloneBucket, fileName1)).To(Equal("FILE1"))
			Expect(GetFileContentsFromBucket(cloneRegion, cloneBucket, fileName2)).To(Equal("FILE2"))
		})
	})

	Context("When the bucket is not versioned", func() {
		BeforeEach(func() {
			backuperInstanceWithUnversionedBucket = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "versioned-backuper-unversioned-bucket",
				Index:      "0",
			}

			unversionedRegion = MustHaveEnv("AWS_TEST_UNVERSIONED_BUCKET_REGION")
			unversionedBucket = MustHaveEnv("AWS_TEST_UNVERSIONED_BUCKET_NAME")
			backuperInstanceWithUnversionedBucket.RunSuccessfully("mkdir -p " + artifactDirPath)

			DeleteAllVersionsFromBucket(unversionedRegion, unversionedBucket)
		})

		AfterEach(func() {
			DeleteAllVersionsFromBucket(unversionedRegion, unversionedBucket)
			backuperInstanceWithUnversionedBucket.RunSuccessfully("rm -rf " + artifactDirPath)
		})

		It("fails with an appropriate error", func() {
			session := backuperInstanceWithUnversionedBucket.Run("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/backup")

			Expect(session).To(gexec.Exit(1))
			Expect(session.Out.Contents()).To(ContainSubstring("is not versioned"))
			Expect(backuperInstanceWithUnversionedBucket.Run("stat " + artifactDirPath + "/blobstore.json")).To(gexec.Exit(1))
		})
	})

	Context("When it connects to a blobstore with custom CA cert", func() {
		BeforeEach(func() {
			backuperInstanceWithCustomCaCertBlobstore = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "unversioned-custom-ca-cert-backuper",
				Index:      "0",
			}
		})

		It("connects and fails with the correct error", func() {
			session := backuperInstanceWithCustomCaCertBlobstore.Run("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
				" /var/vcap/jobs/s3-versioned-blobstore-backup-restorer/bin/bbr/backup")

			Expect(session).To(gexec.Exit(1))
			Expect(session.Out.Contents()).NotTo(ContainSubstring("CERTIFICATE_VERIFY_FAILED"))
			Expect(session.Out.Contents()).NotTo(ContainSubstring("no such host"))
			Expect(session.Out.Contents()).To(ContainSubstring("A header you provided implies functionality that is not implemented"))
		})
	})
})
