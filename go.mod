module github.com/chanzuckerberg/czid-cli

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.2.1
	github.com/aws/aws-sdk-go-v2/credentials v1.1.1
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go-v2/service/s3 v1.2.1
	github.com/aws/smithy-go v1.2.0
	github.com/cheggaaa/pb/v3 v3.0.6
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	golang.org/x/sys v0.21.0 // indirect
)

replace github.com/aws/aws-sdk-go-v2/feature/s3/manager => github.com/chanzuckerberg/aws-sdk-go-v2/feature/s3/manager v1.1.0
