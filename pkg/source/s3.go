package source

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var _ Source = (*s3Source)(nil)

type (
	S3Config struct {
		Endpoint         string `yaml:"endpoint,omitempty" json:"endpoint,omitempty,optional"`
		Region           string `yaml:"region,omitempty" json:"region,omitempty,optional"`
		AccessKeyID      string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty,optional"`
		AccessKeySecret  string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty,optional"`
		S3ForcePathStyle bool   `yaml:"s3ForcePathStyle,omitempty" json:"s3ForcePathStyle,omitempty,optional,default=true"`
		Token            string `yaml:"token,omitempty" json:"token,omitempty,optional"`
		Bucket           string `yaml:"bucket,omitempty" json:"bucket,omitempty,optional"`
		Key              string `yaml:"key,omitempty" json:"key,omitempty,optional"`
	}

	s3Source struct {
		c   *Config
		obj *s3.GetObjectOutput
	}
)

func newS3Source(c *Config) Source {
	return &s3Source{
		c: c,
	}
}

func (s *s3Source) Name() string {
	return s.c.S3.String()
}

func (s *s3Source) Open() error {
	awsConfig := &aws.Config{
		Region:           aws.String(s.c.S3.Region),
		Endpoint:         aws.String(s.c.S3.Endpoint),
		S3ForcePathStyle: aws.Bool(s.Config().S3.S3ForcePathStyle),
	}

	if s.c.S3.AccessKeyID != "" || s.c.S3.AccessKeySecret != "" || s.c.S3.Token != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(s.c.S3.AccessKeyID, s.c.S3.AccessKeySecret, s.c.S3.Token)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return err
	}

	svc := s3.New(sess)

	obj, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.c.S3.Bucket),
		Key:    aws.String(strings.TrimLeft(s.c.S3.Key, "/")),
	})
	if err != nil {
		return err
	}

	s.obj = obj

	return nil
}

func (s *s3Source) Config() *Config {
	return s.c
}

func (s *s3Source) Size() (int64, error) {
	return *s.obj.ContentLength, nil
}

func (s *s3Source) Read(p []byte) (int, error) {
	return s.obj.Body.Read(p)
}

func (s *s3Source) Close() error {
	return s.obj.Body.Close()
}

func (c *S3Config) String() string {
	return fmt.Sprintf("s3 %s:%s %s/%s", c.Region, c.Endpoint, c.Bucket, c.Key)
}
