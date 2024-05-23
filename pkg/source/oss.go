package source

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

var _ Source = (*ossSource)(nil)

type (
	OSSConfig struct {
		Endpoint        string `yaml:"endpoint,omitempty" json:"endpoint,omitempty,optional"`
		AccessKeyID     string `yaml:"accessKeyID,omitempty" json:"accessKeyID,omitempty,optional"`
		AccessKeySecret string `yaml:"accessKeySecret,omitempty" json:"accessKeySecret,omitempty,optional"`
		Bucket          string `yaml:"bucket,omitempty" json:"bucket,omitempty,optional"`
		Key             string `yaml:"key,omitempty" json:"key,omitempty,optional"`
	}

	ossSource struct {
		c      *Config
		cli    *oss.Client
		bucket *oss.Bucket
		r      io.ReadCloser
	}
)

func newOSSSource(c *Config) Source {
	return &ossSource{
		c: c,
	}
}

func (s *ossSource) Name() string {
	return s.c.OSS.String()
}

func (s *ossSource) Open() error {
	cli, err := oss.New(s.c.OSS.Endpoint, s.c.OSS.AccessKeyID, s.c.OSS.AccessKeySecret)
	if err != nil {
		return err
	}

	bucket, err := cli.Bucket(s.c.OSS.Bucket)
	if err != nil {
		return err
	}

	r, err := bucket.GetObject(strings.TrimLeft(s.c.OSS.Key, "/"))
	if err != nil {
		return err
	}

	s.cli = cli
	s.bucket = bucket
	s.r = r

	return nil
}

func (s *ossSource) Config() *Config {
	return s.c
}

func (s *ossSource) Size() (int64, error) {
	meta, err := s.bucket.GetObjectMeta(strings.TrimLeft(s.c.OSS.Key, "/"))
	if err != nil {
		return 0, err
	}
	contentLength := meta.Get("Content-Length")
	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func (s *ossSource) Read(p []byte) (int, error) {
	return s.r.Read(p)
}

func (s *ossSource) Close() error {
	return s.r.Close()
}

func (c *OSSConfig) String() string {
	return fmt.Sprintf("oss %s %s/%s", c.Endpoint, c.Bucket, c.Key)
}
