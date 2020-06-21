package config

// S3 конфиг для подключения к S3 хранилищу
type S3 struct {
	Key          string
	Secret       string
	Region       string
	Bucket       string
	Endpoint     string
	BackupFolder string
}

// NewS3 конструктор
func NewS3(key, secret, region, bucket, endpoint, backupFolder string) *S3 {
	return &S3{
		Key:          key,
		Secret:       secret,
		Region:       region,
		Bucket:       bucket,
		Endpoint:     endpoint,
		BackupFolder: backupFolder,
	}
}
