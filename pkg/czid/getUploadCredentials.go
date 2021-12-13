package czid

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type getUploadCredentialsReq struct{}

type getUploadCredentialsRes struct {
	AccessKeyID     string    `json:"access_key_id"`
	Expiration      time.Time `json:"expiration"`
	SecretAccessKey string    `json:"secret_access_key"`
	SessionToken    string    `json:"session_token"`
}

func (c *Client) GetUploadCredentials(sampleID int) (aws.Credentials, error) {
	var res getUploadCredentialsRes
	err := c.request(
		"GET",
		fmt.Sprintf("/samples/%d/upload_credentials", sampleID),
		"",
		getUploadCredentialsReq{},
		&res,
	)

	return aws.Credentials{
		AccessKeyID:     res.AccessKeyID,
		Expires:         res.Expiration,
		SecretAccessKey: res.SecretAccessKey,
		SessionToken:    res.SessionToken,
	}, err
}
