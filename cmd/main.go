package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"go-db-backup-to-s3/config"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const openFileOptions = os.O_CREATE | os.O_RDWR
const openFilePermissions os.FileMode = 0660

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
	return config.NewMySql(viper.GetString("db.name"), viper.GetString("db.user"), viper.GetString("db.password"))
}

// initMysqlDumpConfig инициализирует конфиг mysqldump
func initMysqlDumpConfig() *config.MySqlDump {
	return config.NewMySqlDump(viper.GetString("dump.ignoreTable"), viper.GetBool("dump.addDropTable"))
}

// initMysqlDumpConfig инициализирует конфиг mysqldump
func initBackupConfig() *config.Backup {
	return config.NewBackup(viper.GetString("backup.folder"), viper.GetString("backup.name"), viper.GetString("backup.extension"))
}

// generateBackupDate генерирует дату бекапа
func generateBackupDate() string {
	dt := time.Now()
	return strconv.Itoa(dt.Year()) + "-" + dt.Weekday().String()
}

// gzipFile сжимает файл в архив
func gzipFile(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}

	filename := filepath.Base(source)
	target = filepath.Join(target, fmt.Sprintf("%s.gz", filename))
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

	backupDate := generateBackupDate()
	backupFullPath := backupConfig.Folder + backupConfig.Name + "." + backupDate + backupConfig.Extension
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

	io.Copy(outfile, stdout)

	fmt.Println("finish dumping")

	err = gzipFile(backupFullPath, "/var/www/backups/gzipped")
	if err != nil {
		fmt.Println("error while gzip file")
	}

	fmt.Println("finish gzip")

	err = deleteFile(backupFullPath)
	if err != nil {
		fmt.Println("error while delete file")
	}

	fmt.Println("finish delete file")
}
