package internal

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"go-db-backup-to-s3/internal/types"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// AclPrivate приватный доступ к файлу на S3
const AclPrivate = "private"

// AclPublic публичный доступ к фалу на S3
const AclPublic = "public"

// S3Client клиент для работы с S3
type S3Client struct {
	Config *types.S3
	Client *s3.S3
}

// NewS3Client конструктор
func NewS3Client(
	config *types.S3,
) *S3Client {
	client := &S3Client{
		Config: config,
	}
	client.initClient()
	return client
}

// initClient инициализация клиента
func (c *S3Client) initClient() {
	newSession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(c.Config.Key, c.Config.Secret, ""),
		Endpoint:    aws.String(c.Config.Endpoint),
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		log.Fatal(err)
	}
	c.Client = s3.New(newSession)
}

// generateS3FileName генерирует путь и название файла для S3
func (c *S3Client) generateS3FileName(backupS3Folder, source string) string {
	return backupS3Folder + filepath.Base(source)
}

// UploadFile загружает файл на S3
func (c *S3Client) UploadFile(fileName string, private bool) error {
	privateMod := AclPublic
	if private == true {
		privateMod = AclPrivate
	}
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	_, _ = file.Read(buffer)

	object := s3.PutObjectInput{
		Bucket:        aws.String(c.Config.Bucket),
		Key:           aws.String(c.generateS3FileName(c.Config.BackupFolder, fileName)),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		ACL:           aws.String(privateMod),
	}
	_, err = c.Client.PutObject(&object)
	return err
}

// GetPresignLink генерирует ссылку для скачивания приватного файла
func (c *S3Client) GetPresignLink(fileName string, time time.Duration) (string, error) {
	req, _ := c.Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.Config.Bucket),
		Key:    aws.String(c.generateS3FileName(c.Config.BackupFolder, fileName)),
	})

	url, err := req.Presign(time)
	if err != nil {
		return "", err
	}
	return url, nil
}
