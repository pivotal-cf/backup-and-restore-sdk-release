package s3bucket_test

import (
	"s3-blobstore-backup-restore/s3bucket"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating an S3 Client", func() {

	When("we are not using an IAMProfile", func() {
		It("provides static credentials", func() {

			s3Object, err := s3bucket.NewS3ClientImpl("", "", s3bucket.AccessKey{"user", "pass"}, false, false)

			Expect(err).NotTo(HaveOccurred())
			creds, err := s3Object.Client.Config.Credentials.Get()
			Expect(err).NotTo(HaveOccurred())
			Expect(creds.ProviderName).To(Equal("StaticProvider"))
		})
	})

	When("we are using an IAMProfile", func() {
		It("provides EC2 Role credentials", func() {
			roleCredentials := &credentials.Credentials {}

			s3bucket.SetCredIAMProvider(func(c client.ConfigProvider, options ...func(*ec2rolecreds.EC2RoleProvider)) *credentials.Credentials  {
				return roleCredentials
			})

			s3Object, err := s3bucket.NewS3ClientImpl("", "", s3bucket.AccessKey{"user", "pass"}, true, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(s3Object.Client.Config.Credentials).To(BeIdenticalTo(roleCredentials))
		})
	})

	When("we want to use a path style bucket addresses", func() {
		It("adds the appropriate property to the config object", func() {

			s3Object, err := s3bucket.NewS3ClientImpl("", "", s3bucket.AccessKey{}, false, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(s3Object.Client.Config.S3ForcePathStyle).To(Equal(aws.Bool(true)))
		})
	})

	When("we want to use a v-host style bucket addresses", func(){
		It("adds the appropriate property to the config object", func() {

			s3Object, err := s3bucket.NewS3ClientImpl("", "", s3bucket.AccessKey{}, false, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(s3Object.Client.Config.S3ForcePathStyle).To(Equal(aws.Bool(false)))
		})
	})
})
