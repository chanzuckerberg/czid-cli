package upload

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/chanzuckerberg/idseq-cli-v2/pkg/util"
	"github.com/cheggaaa/pb/v3"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type partChannelClient struct {
	aws.HTTPClient
	parts chan int64
}

func (c *partChannelClient) Do(r *http.Request) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(r)
	partNumber := r.URL.Query().Get("partNumber")
	if resp.StatusCode < 400 && r.Method == "PUT" && partNumber != "" {
		partNumber, _ := strconv.ParseInt(partNumber, 10, 64)
		c.parts <- partNumber
	}
	return resp, err
}

type Uploader struct {
	u *manager.Uploader
	c *partChannelClient
}

func NewUploader(creds aws.Credentials) Uploader {
	provider := credentials.StaticCredentialsProvider{
		Value: creds,
	}

	var pC partChannelClient
	client := s3.New(s3.Options{}, func(o *s3.Options) {
		pC = partChannelClient{HTTPClient: o.HTTPClient, parts: make(chan int64)}
		o.HTTPClient = &pC
		o.Credentials = provider
		o.Region = "us-west-2"
	})
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.LeavePartsOnError = true
	})
	return Uploader{u: uploader, c: &pC}
}

func (u *Uploader) runProgressBar(fileSize int64) {
	var minPartNumber *int64
	count := int64(0)
	var bar *pb.ProgressBar
	for partNumber := range u.c.parts {
		if minPartNumber == nil {
			// part numbers start at 1
			m := partNumber - 1
			minPartNumber = &m
			bar = pb.Full.Start64(fileSize)
			bar.Set(pb.Bytes, true)
			bar.SetCurrent(m * u.u.PartSize)
		}
		if partNumber > *minPartNumber {
			count += 1
		}
		bar.SetCurrent((*minPartNumber + count) * u.u.PartSize)
	}
	if bar != nil {
		(*bar).Finish()
	}
}

func (u *Uploader) UploadFile(filename string, s3path string, multipartUploadId *string) error {
	u.c.parts = make(chan int64)
	defer close(u.c.parts)

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	go u.runProgressBar(stat.Size())

	parsedPath, err := url.Parse(s3path)
	if err != nil {
		return err
	}

	key := util.TrimLeadingSlash(parsedPath.Path)

	input := s3.PutObjectInput{
		Bucket: &parsedPath.Host,
		Key:    &key,
		Body:   f,
	}

	if multipartUploadId != nil {
		fmt.Printf("resuming upload of %s\n", filename)
		_, err = u.u.ResumeUpload(context.Background(), &input, multipartUploadId)
		if err != nil {
			fmt.Println("could not resume upload, starting fresh upload")
			close(u.c.parts)
			u.c.parts = make(chan int64)
			defer close(u.c.parts)
			go u.runProgressBar(stat.Size())
			_, err = u.u.Upload(context.Background(), &input)
		}
	} else {
		fmt.Printf("starting upload of %s\n", filename)
		_, err = u.u.Upload(context.Background(), &input)
	}
	return err
}
