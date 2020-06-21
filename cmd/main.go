package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"
	"go-db-backup-to-s3/config"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func doExample() {
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

// initMysqlConfig инициализирует конфиг MySQL
func initMysqlConfig() *config.MySql {
	return config.NewMySql(
		viper.GetString("db.name"),
		viper.GetString("db.user"),
		viper.GetString("db.password"),
	)
}

// initMysqlDumpConfig инициализирует конфиг mysqldump
func initMysqlDumpConfig() *config.MySqlDump {
	return config.NewMySqlDump(
		viper.GetString("dump.ignoreTable"),
		viper.GetBool("dump.addDropTable"),
	)
}

// initMysqlDumpConfig инициализирует конфиг mysqldump
func initBackupConfig() *config.Backup {
	return config.NewBackup(
		viper.GetString("backup.folder"),
		viper.GetString("backup.fileName"),
		viper.GetString("backup.fileExtension"),
		viper.GetString("backup.gzipExtension"),
	)
}

// initS3Config инициализирует конфиг S3
func initS3Config() *config.S3 {
	return config.NewS3(
		viper.GetString("s3.key"),
		viper.GetString("s3.secret"),
		viper.GetString("s3.region"),
		viper.GetString("s3.bucket"),
		viper.GetString("s3.endpoint"),
		viper.GetString("s3.backupFolder"),
	)
}

// initTelegramConfig инициализирует конфиг Telegram
func initTelegramConfig() *config.Telegram {
	chatIdsFromConfig := viper.GetIntSlice("telegram.chatIds")
	chatIds := make([]int64, 0, len(chatIdsFromConfig))
	for _, chatId := range chatIdsFromConfig {
		chatIds = append(chatIds, int64(chatId))
	}
	return config.NewTelegramConfig(
		viper.GetString("telegram.apiToken"),
		chatIds,
	)
}

// generateBackupDate генерирует дату бекапа
func generateBackupDate() string {
	dt := time.Now()
	return strconv.Itoa(dt.Year()) + "-" + dt.Weekday().String()
}

// gzipFile сжимает файл в архив
func gzipFile(source, gzipExtension string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target := source + gzipExtension
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = filename
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)
	return err
}

// deleteFile удаляет файл
func deleteFile(source string) error {
	err := os.Remove(source)
	return err
}

// generateS3FileName генерирует путь и название файла для S3
func generateS3FileName(backupS3Folder, source string) string {
	return backupS3Folder + filepath.Base(source)
}

func main() {
	viper.AddConfigPath("./config/")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	mysqlConfig := initMysqlConfig()
	mysqlDumpConfig := initMysqlDumpConfig()
	backupConfig := initBackupConfig()
	s3config := initS3Config()
	telegramConfig := initTelegramConfig()

	backupDate := generateBackupDate()
	backupFullPath := backupConfig.Folder + backupConfig.FileName + "." + backupDate + backupConfig.BackupExtension
	backupGzipFullPath := backupFullPath + backupConfig.GzipExtension
	fmt.Println(backupFullPath)

	mysqlDumpExtras := "--ignore-table=" + mysqlDumpConfig.IgnoreTable + " --add-drop-table"
	cmd := exec.Command(
		"mysqldump",
		"-u"+mysqlConfig.User,
		"-p"+mysqlConfig.Password,
		mysqlConfig.Name,
		mysqlDumpExtras,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	outfile, err := os.Create(backupFullPath)
	if err != nil {
		log.Fatal(err)
	}
	defer outfile.Close()

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// read command's stdout line by line
	in := bufio.NewWriter(outfile)
	defer in.Flush()

	if _, err := io.Copy(outfile, stdout); err != nil {
		log.Fatal(err)
	}

	fmt.Println("finish dumping")

	err = gzipFile(backupFullPath, backupConfig.GzipExtension)
	if err != nil {
		fmt.Println("error while gzip file")
	}

	fmt.Println("finish gzip")

	err = deleteFile(backupFullPath)
	if err != nil {
		fmt.Println("error while delete file")
	}

	fmt.Println("finish delete file")

	newSession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(s3config.Key, s3config.Secret, ""),
		Endpoint:    aws.String(s3config.Endpoint),
		Region:      aws.String("us-east-1"),
	})
	if err != nil {
		fmt.Println(err)
	}
	s3Client := s3.New(newSession)

	file, err := os.Open(backupGzipFullPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	// Get file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)
	_, _ = file.Read(buffer)

	object := s3.PutObjectInput{
		Bucket:        aws.String(s3config.Bucket),
		Key:           aws.String(generateS3FileName(s3config.BackupFolder, backupGzipFullPath)),
		Body:          bytes.NewReader(buffer),
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(http.DetectContentType(buffer)),
		ACL:           aws.String("private"),
	}
	_, err = s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println("finish upload")

	err = deleteFile(backupGzipFullPath)
	if err != nil {
		fmt.Println("error while delete file")
	}

	fmt.Println("finish delete gzip file")

	req, _ := s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s3config.Bucket),
		Key:    aws.String(generateS3FileName(s3config.BackupFolder, backupGzipFullPath)),
	})

	urlStr, err := req.Presign(24 * time.Hour)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(telegramConfig.ChatIds)
	bot, err := tgbotapi.NewBotAPI(telegramConfig.ApiToken)
	if err != nil {
		log.Panic(err)
	}
	for _, chatId := range telegramConfig.ChatIds {
		msg := tgbotapi.NewMessage(chatId, "db backup of "+mysqlConfig.Name+" finished\nDownload file:\n"+urlStr)
		_, _ = bot.Send(msg)
	}
}
