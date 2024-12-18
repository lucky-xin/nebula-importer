package source

type (
	Config struct {
		Local *LocalConfig `yaml:"local,omitempty" json:"local,omitempty,optional"`
		S3    *S3Config    `yaml:"s3,omitempty" json:"s3,omitempty,optional"`
		OSS   *OSSConfig   `yaml:"oss,omitempty" json:"oss,omitempty,optional"`
		FTP   *FTPConfig   `yaml:"ftp,omitempty" json:"ftp,omitempty,optional"`
		SFTP  *SFTPConfig  `yaml:"sftp,omitempty" json:"sftp,omitempty,optional"`
		HDFS  *HDFSConfig  `yaml:"hdfs,omitempty" json:"hdfs,omitempty,optional"`
		GCS   *GCSConfig   `yaml:"gcs,omitempty" json:"gcs,omitempty,optional"`
		SQL   *SQLConfig   `yaml:"sql,omitempty" json:"sql,omitempty,optional"`
		// The following is format information
		CSV *CSVConfig `yaml:"csv,omitempty" json:"csv,omitempty,optional"`
	}

	CSVConfig struct {
		Delimiter  string `yaml:"delimiter,omitempty" json:"delimiter,omitempty,optional"`
		Comment    string `yaml:"comment,omitempty" json:"comment,omitempty,optional"`
		WithHeader bool   `yaml:"withHeader,omitempty" json:"withHeader,omitempty,optional"`
		LazyQuotes bool   `yaml:"lazyQuotes,omitempty" json:"lazyQuotes,omitempty,optional"`
	}
)

func (c *Config) Clone() *Config {
	cpy := *c
	switch {
	case cpy.S3 != nil:
		cpy1 := *cpy.S3
		cpy.S3 = &cpy1
	case cpy.OSS != nil:
		cpy1 := *cpy.OSS
		cpy.OSS = &cpy1
	case cpy.FTP != nil:
		cpy1 := *cpy.FTP
		cpy.FTP = &cpy1
	case cpy.SFTP != nil:
		cpy1 := *cpy.SFTP
		cpy.SFTP = &cpy1
	case cpy.HDFS != nil:
		cpy1 := *cpy.HDFS
		cpy.HDFS = &cpy1
	case cpy.GCS != nil:
		cpy1 := *cpy.GCS
		cpy.GCS = &cpy1
	case cpy.SQL != nil:
		cpy1 := *cpy.SQL
		cpy.SQL = &cpy1
	default:
		cpy1 := *cpy.Local
		cpy.Local = &cpy1
	}
	return &cpy
}
