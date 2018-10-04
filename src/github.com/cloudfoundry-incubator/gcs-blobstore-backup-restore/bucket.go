package gcs

import (
	"fmt"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const readWriteScope = "https://www.googleapis.com/auth/devstorage.read_write"

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	VersioningEnabled() (bool, error)
	ListBlobs() ([]Blob, error)
	CopyVersion(blob Blob, sourceBucketName string) error
	LastBackupPrefix() (string, error)
	ListLastBackupBlobs() ([]Blob, error)
	ListPenultimateBackupBlobs() ([]Blob, error)
	CopyBlob(destinationBucket Bucket, prefix string, blob Blob) error
}

type BucketPair struct {
	liveBucket   Bucket
	backupBucket Bucket
}

func BuildBuckets(config map[string]Config) (map[string]BucketPair, error) {
	buckets := map[string]BucketPair{}

	var err error
	for bucketPairId, bucketConfig := range config {
		var liveBucket Bucket
		var backupBucket Bucket
		liveBucket, err = NewSDKBucket(bucketConfig.ServiceAccountKey, bucketConfig.LiveBucketName, bucketPairId)
		if err != nil {
			return nil, err
		}
		backupBucket, err = NewSDKBucket(bucketConfig.ServiceAccountKey, bucketConfig.BackupBucketName, bucketPairId)
		if err != nil {
			return nil, err
		}

		buckets[bucketPairId] = BucketPair{liveBucket: liveBucket, backupBucket: backupBucket}

	}

	return buckets, nil
}

type Blob struct {
	Name         string `json:"name"`
	Prefix       string
	GenerationID int64 `json:"generation_id"`
}

func (b Blob) Path() string {
	if b.Prefix != "" {
		return b.Prefix + "/" + b.Name
	}
	return b.Name
}

type BucketBackup struct {
	LiveBucketName   string `json:"live_bucket_name"`
	BackupBucketName string `json:"backup_bucket_name"`
	Blobs            []Blob `json:"blobs"`
}

type SDKBucket struct {
	name       string
	handle     *storage.BucketHandle
	ctx        context.Context
	client     *storage.Client
	identifier string
}

func (b SDKBucket) CopyBlob(destinationBucket Bucket, prefix string, blob Blob) error {
	ctx := context.Background()

	sourceObjectHandle := b.handle.Object(blob.Path())
	destinationName := fmt.Sprintf("%s/%s", prefix, blob.Name)

	copier := b.client.Bucket(destinationBucket.Name()).Object(destinationName).CopierFrom(sourceObjectHandle)
	_, err := copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("error copying blob 'gs://%s/%s': %s", b.name, blob.Path(), err)
	}

	return nil
}

func (b SDKBucket) LastBackupPrefix() (string, error) {
	var directories []string
	objectsIterator := b.handle.Objects(b.ctx, &storage.Query{
		Delimiter: "/",
		Prefix:    "",
	})
	for {
		objectAttributes, err := objectsIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return "", err
		}

		fmt.Printf("found last backup directory: '%s', prefix: '%s'\n", objectAttributes.Name, objectAttributes.Prefix)

		directories = append(directories, objectAttributes.Prefix)
	}

	if len(directories) == 0 {
		fmt.Println("no last backup directory found")
		return "", nil
	}
	latestDirectory := directories[len(directories)-1]

	return path.Join(latestDirectory, b.identifier), nil
}

func (b SDKBucket) ListLastBackupBlobs() ([]Blob, error) {
	prefix, err := b.LastBackupPrefix()
	if err != nil {
		return nil, err
	}
	fmt.Printf("Last backup prefix: '%s'\n", prefix)

	if prefix == "" {
		return []Blob{}, nil
	}

	var blobs []Blob
	objectsIterator := b.handle.Objects(b.ctx, &storage.Query{
		Delimiter: prefix,
		Prefix:    prefix,
	})
	for {
		objectAttributes, err := objectsIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		fmt.Printf("found last backup blob: '%s', prefix: '%s'\n", objectAttributes.Name, objectAttributes.Prefix)
		name := strings.TrimPrefix(objectAttributes.Name, prefix+"/")

		blobs = append(blobs, Blob{Name: name, Prefix: prefix, GenerationID: objectAttributes.Generation})
	}

	return blobs, nil
}

func (b SDKBucket) ListPenultimateBackupBlobs() ([]Blob, error) {
	list, err := b.ListBlobs()
	if err != nil {
		return nil, err
	}

	prefix := b.getPrefixFromAgesAgo(list)
	if prefix == "" {
		return []Blob{}, nil
	}

	var blobs []Blob
	objectsIterator := b.handle.Objects(b.ctx, &storage.Query{
		Delimiter: prefix,
		Prefix:    prefix,
	})
	for {
		objectAttributes, err := objectsIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		fmt.Printf("found last backup blob: '%s', prefix: '%s'\n", objectAttributes.Name, objectAttributes.Prefix)
		name := strings.TrimPrefix(objectAttributes.Name, prefix+"/")

		blobs = append(blobs, Blob{Name: name, Prefix: prefix, GenerationID: objectAttributes.Generation})
	}

	return blobs, nil
}

func (b SDKBucket) getPrefixFromAgesAgo(list []Blob) string {
	if len(list) == 0 {
		return ""
	}

	latestBlob := list[len(list)-1]
	parts := strings.Split(latestBlob.Path(), b.identifier)
	latestPrefix := fmt.Sprintf("%s%s", parts[0], b.identifier)

	for i := len(list) - 2; i >= 0; i-- {
		blob := list[i]
		if !strings.HasPrefix(blob.Name, latestPrefix) {
			parts := strings.Split(blob.Path(), b.identifier)
			return fmt.Sprintf("%s%s", parts[0], b.identifier)
		}
	}

	return ""
}

func NewSDKBucket(serviceAccountKeyJson, name, identifier string) (SDKBucket, error) {
	ctx := context.Background()

	creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountKeyJson), readWriteScope)
	if err != nil {
		return SDKBucket{}, err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return SDKBucket{}, err
	}

	handle := client.Bucket(name)

	return SDKBucket{name: name, identifier: identifier, handle: handle, ctx: ctx, client: client}, nil
}

func (b SDKBucket) Name() string {
	return b.name
}

func (b SDKBucket) VersioningEnabled() (bool, error) {
	attrs, err := b.handle.Attrs(b.ctx)
	if err != nil {
		return false, err
	}

	return attrs.VersioningEnabled, nil
}

func (b SDKBucket) ListBlobs() ([]Blob, error) {
	var blobs []Blob

	objectsIterator := b.handle.Objects(b.ctx, nil)
	for {
		objectAttributes, err := objectsIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		blobs = append(blobs, Blob{Name: objectAttributes.Name, GenerationID: objectAttributes.Generation})
	}

	return blobs, nil
}

func (b SDKBucket) CopyVersion(blob Blob, sourceBucketName string) error {
	ctx := context.Background()

	sourceObjectHandle := b.client.Bucket(sourceBucketName).Object(blob.Name)
	_, err := sourceObjectHandle.Generation(blob.GenerationID).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("error getting blob version attributes 'gs://%s/%s#%d': %s", sourceBucketName, blob.Path(), blob.GenerationID, err)
	}

	if b.name == sourceBucketName {
		attrs, err := b.handle.Object(blob.Name).Attrs(ctx)

		if err == nil && attrs.Generation == blob.GenerationID {
			return nil
		}
	}

	source := sourceObjectHandle.Generation(blob.GenerationID)
	copier := b.handle.Object(blob.Name).CopierFrom(source)
	_, err = copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("error copying blob 'gs://%s/%s#%d': %s", sourceBucketName, blob.Path(), blob.GenerationID, err)
	}

	return nil
}
