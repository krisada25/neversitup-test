package config

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/storage"

	"google.golang.org/api/option"
)

var FileManage *FileConnect

type FileConnect struct {
	Ctx        context.Context
	Client     *storage.Client
	BucketName string
	Bucket     *storage.BucketHandle
}

func FileHandler() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile("credential/gcpkey.json"))
	if err != nil {
		fmt.Println(err)
	}
	bucket := os.Getenv("GCS_BUCKET")
	bkt := client.Bucket(bucket)
	d := &FileConnect{
		Ctx:        ctx,
		Client:     client,
		Bucket:     bkt,
		BucketName: bucket,
	}
	FileManage = d
}
