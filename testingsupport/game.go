package testingsupport

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type S3 struct {
	s3iface.S3API
	GetObjectOutput *s3.GetObjectOutput
	Err             error
}

func (s *S3) GetObject(*s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	return s.GetObjectOutput, s.Err
}
