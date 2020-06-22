package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go-db-backup-to-s3/internal"
	"go-db-backup-to-s3/internal/types"
	"os"
	"strings"
	"time"
)

// initMysqlConfig инициализирует конфиг MySQL
func initMysqlConfig() *types.MySql {
	return types.NewMySql(
		viper.GetString("db.host"),
		viper.GetString("db.port"),
		viper.GetString("db.name"),
		viper.GetString("db.user"),
		viper.GetString("db.password"),
	)
}

// initMysqlDumpConfig инициализирует конфиг mysqldump
func initMysqlDumpConfig() *types.MySqlDump {
	return types.NewMySqlDump(
		viper.GetString("dump.ignoreTable"),
		viper.GetBool("dump.addDropTable"),
	)
}

// initBackupConfig инициализирует конфиг для бекапа
func initBackupConfig() *types.Backup {
	return types.NewBackup(
		viper.GetString("backup.folder"),
		viper.GetString("backup.backupExtension"),
	)
}

// initS3Config инициализирует конфиг S3
func initS3Config() *types.S3 {
	return types.NewS3(
		viper.GetString("s3.key"),
		viper.GetString("s3.secret"),
		viper.GetString("s3.region"),
		viper.GetString("s3.bucket"),
		viper.GetString("s3.endpoint"),
		viper.GetString("s3.backupFolder"),
	)
}

// initTelegramConfig инициализирует конфиг Telegram
func initTelegramConfig() *types.Telegram {
	chatIdsFromConfig := viper.GetIntSlice("telegram.chatIds")
	chatIds := make([]int64, 0, len(chatIdsFromConfig))
	for _, chatId := range chatIdsFromConfig {
		chatIds = append(chatIds, int64(chatId))
	}
	return types.NewTelegramConfig(
		viper.GetString("telegram.apiToken"),
		chatIds,
	)
}

func main() {
	viper.AddConfigPath("./cmd/config/")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("error reading configs: %s\n", err)
	}

	mysqlDumpConfig := initMysqlDumpConfig()
	backupConfig := initBackupConfig()
	s3config := initS3Config()
	telegramConfig := initTelegramConfig()

	dir, err := os.Open("./cmd/config/databases") // читаем текущий путь
	if err != nil {
		fmt.Printf("error opening dir: %s\n", err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1) // читаем текущую директорию
	if err != nil {
		fmt.Printf("error reading dir: %s\n", err)
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.Name() != ".gitkeep" {
			fmt.Println(fileInfo.Name())
			file := strings.Split(fileInfo.Name(), ".")
			viper.AddConfigPath("./cmd/config/databases")
			viper.SetConfigName(file[0])
			viper.SetConfigType("yaml")
			err := viper.ReadInConfig()
			if err != nil {
				fmt.Printf("error reading db config: %s\n%s\n", fileInfo.Name(), err)
			}
			mysqlConfig := initMysqlConfig()
			dumper := internal.NewDumper(backupConfig, mysqlDumpConfig, mysqlConfig)
			err = dumper.DumpDb()
			if err != nil {
				fmt.Printf("error in dumper: %s\n", err)
			}

			gzipper := internal.NewGzipper()
			err = gzipper.GzipFile(dumper.FileName)
			if err != nil {
				fmt.Printf("error while gzip file: %s\n", err)
			}

			deleter := internal.NewDeleter()
			err = deleter.DeleteFile(dumper.FileName)
			if err != nil {
				fmt.Printf("error while delete file: %s\n", err)
			}

			s3client := internal.NewS3Client(s3config)
			err = s3client.UploadFile(gzipper.FileName, true)
			if err != nil {
				fmt.Printf("error while uploading file to S3: %s\n", err)
			}

			err = deleter.DeleteFile(gzipper.FileName)
			if err != nil {
				fmt.Printf("error while deleting gz file: %s\n", err)
			}

			presignLink, err := s3client.GetPresignLink(gzipper.FileName, 24*time.Hour)
			if err != nil {
				fmt.Printf("error while generating presign link: %s\n", err)
			}

			tgClient := internal.NewTgClient(telegramConfig)
			tgClient.SendDbBackupFinishMessage(mysqlConfig.Name, presignLink)
		}
	}
}
