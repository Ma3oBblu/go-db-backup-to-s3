package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"os"
	"strings"
)

func main() {
	key := os.Getenv("SPACES_KEY")
	secret := os.Getenv("SPACES_SECRET")

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:    aws.String("https://nyc3.digitaloceanspaces.com"),
		Region:      aws.String("us-east-1"),
	}

	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		Bucket: aws.String("example-space-name"),
		Key:    aws.String("file.ext"),
		Body:   strings.NewReader("The contents of the file."),
		ACL:    aws.String("private"),
	}
	_, err := s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String("example-space-name"),
		Key:    aws.String("file.ext"),
	}

	result, err := s3Client.GetObject(input)
	if err != nil {
		fmt.Println(err.Error())
	}

	out, err := os.Create("/tmp/local-file.ext")
	defer out.Close()

	_, err = io.Copy(out, result.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
}
