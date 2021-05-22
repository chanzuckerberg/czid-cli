package idseq

import (
	"encoding/json"
	"net/url"
    "fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type getUploadCredentialsReq struct {}

type getUploadCredentialsRes struct {
	AccessKeyID     string    `json:"access_key_id"`
	Expiration      time.Time `json:"expiration"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
}

func (c *Client) GetUploadCredentials(sampleID int) (credentialsRes, error) {
	var res getUploadCredentialsRes
	err := c.request(
		"GET",
		fmt.Sprintf("/samples/%d/upload_credentials",
		"",
		getUploadCredentialsReq{},
		&res,
	)

	credentials := aws.Credentials{
		AccessKeyID:     res.Credentials.AccessKeyID,
		Expires:         res.Credentials.Expiration,
		SecretAccessKey: res.Credentials.SecretAccessKey,
		SessionToken:    res.Credentials.SessionToken,
	}
	return metadataFields, err
}
