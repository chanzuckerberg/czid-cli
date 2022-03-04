package upload

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"github.com/chanzuckerberg/czid-cli/pkg/util"
	"github.com/cheggaaa/pb/v3"
)

// from s3 manager
const MinUploadPartSize int64 = 1024 * 1024 * 5
const DefaultUploadPartSize = MinUploadPartSize

type partChannelClient struct {
	aws.HTTPClient
	parts chan int64
}

func (c *partChannelClient) Do(r *http.Request) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(r)
	if err != nil {
		return resp, err
	}
	partNumber := r.URL.Query().Get("partNumber")
	if resp.StatusCode < 400 && r.Method == "PUT" && partNumber != "" {
		partNumber, _ := strconv.ParseInt(partNumber, 10, 64)
		c.parts <- partNumber
	}
	return resp, err
}

type Uploader struct {
	u      *manager.Uploader
	c      *partChannelClient
	client *s3.Client
}

func NewUploader(creds aws.Credentials, disableBuffer bool) Uploader {
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
		u.Concurrency = runtime.NumCPU()
		if !disableBuffer {
			u.BufferProvider = manager.NewBufferedReadSeekerWriteToPool(int(DefaultUploadPartSize) * runtime.NumCPU())
		}
	})
	return Uploader{u: uploader, c: &pC, client: client}
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

func (u *Uploader) UploadFiles(filenames []string, s3path string, multipartUploadId *string) error {
	size := int64(0)
	for _, filename := range filenames {
		stat, err := os.Stat(filename)
		if err != nil {
			return err
		}
		size += stat.Size()
	}

	readers := make([]io.Reader, len(filenames))
	for i, filename := range filenames {
		var err error
		readers[i], err = os.Open(filename)
		if err != nil {
			return err
		}
	}

	reader := io.MultiReader(readers...)

	parsedPath, err := url.Parse(s3path)
	if err != nil {
		return err
	}

	key := util.TrimLeadingSlash(parsedPath.Path)

	input := s3.PutObjectInput{
		Bucket: &parsedPath.Host,
		Key:    &key,
		Body:   bufio.NewReader(reader),
	}

	if size <= 1024*1024*5 {
		_, err := u.client.PutObject(context.Background(), &input)
		return err
	}

	_, err = u.client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: &parsedPath.Host,
		Key:    &key,
	})

	if err != nil {
		var nfe smithy.APIError
		// if our permissions give us access to specific resources, and those
		//   resources don't exist we get Forbidden instead of NotFound. This
		//   is how our permissions are by default so Forbidden is required here.
		if !errors.As(err, &nfe) || (nfe.ErrorCode() != "NotFound" && nfe.ErrorCode() != "Forbidden") {
			fmt.Println(nfe.ErrorCode())
			return err
		}
	} else {
		fmt.Printf("skipping upload of %s: already uploaded\n", strings.Join(filenames, ", "))
		return nil
	}

	u.c.parts = make(chan int64)
	defer close(u.c.parts)

	go u.runProgressBar(size)

	if multipartUploadId != nil {
		fmt.Printf("resuming upload of %s\n", strings.Join(filenames, ", "))
		_, err = u.u.ResumeUpload(context.Background(), &input, multipartUploadId)
		if err != nil {
			fmt.Println("could not resume upload, starting fresh upload")
			close(u.c.parts)
			u.c.parts = make(chan int64)
			defer close(u.c.parts)
			go u.runProgressBar(size)
			_, err = u.u.Upload(context.Background(), &input)
		}
	} else {
		fmt.Printf("starting upload of %s\n", strings.Join(filenames, ", "))
		_, err = u.u.Upload(context.Background(), &input)
	}
	return err
}
