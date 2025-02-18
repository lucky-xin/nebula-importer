package source

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/option"
)

var _ Source = (*gcsSource)(nil)

type (
	GCSConfig struct {
		Endpoint              string `yaml:"endpoint,omitempty" json:"endpoint,omitempty,optional"`
		CredentialsFile       string `yaml:"credentialsFile,omitempty" json:"credentialsFile,omitempty,optional"`
		CredentialsJSON       string `yaml:"credentialsJSON,omitempty" json:"credentialsJSON,omitempty,optional"`
		WithoutAuthentication bool   `yaml:"withoutAuthentication,omitempty" json:"withoutAuthentication,omitempty,optional"`
		Bucket                string `yaml:"bucket,omitempty" json:"bucket,omitempty,optional"`
		Key                   string `yaml:"key,omitempty" json:"key,omitempty,optional"`
	}

	gcsSource struct {
		c      *Config
		reader *storage.Reader
	}
)

func newGCSSource(c *Config) Source {
	return &gcsSource{
		c: c,
	}
}

func (s *gcsSource) Name() string {
	return s.c.GCS.String()
}

func (s *gcsSource) Open() error {
	var gcsOptions []option.ClientOption
	if s.c.GCS.Endpoint != "" {
		gcsOptions = append(gcsOptions, option.WithEndpoint(s.c.GCS.Endpoint))
	}

	if s.c.GCS.CredentialsFile != "" {
		gcsOptions = append(gcsOptions, option.WithCredentialsFile(s.c.GCS.CredentialsFile))
	} else if s.c.GCS.CredentialsJSON != "" {
		gcsOptions = append(gcsOptions, option.WithCredentialsJSON([]byte(s.c.GCS.CredentialsJSON)))
	} else if s.c.GCS.WithoutAuthentication {
		gcsOptions = append(gcsOptions, option.WithoutAuthentication())
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, gcsOptions...)
	if err != nil {
		return err
	}
	defer client.Close()

	obj := client.Bucket(s.c.GCS.Bucket).Object(strings.TrimLeft(s.c.GCS.Key, "/"))
	if s.reader, err = obj.NewReader(ctx); err != nil {
		return err
	}
	return nil
}

func (s *gcsSource) Config() *Config {
	return s.c
}

func (s *gcsSource) Size() (int64, error) {
	return s.reader.Attrs.Size, nil
}

func (s *gcsSource) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *gcsSource) Close() error {
	return s.reader.Close()
}

func (c *GCSConfig) String() string {
	return fmt.Sprintf("gcs %s %s/%s", c.Endpoint, c.Bucket, c.Key)
}
