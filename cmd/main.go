package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go-db-backup-to-s3/config"
	"go-db-backup-to-s3/internal"
	"time"
)

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

// initBackupConfig инициализирует конфиг для бекапа
func initBackupConfig() *config.Backup {
	return config.NewBackup(
		viper.GetString("backup.folder"),
		viper.GetString("backup.fileName"),
		viper.GetString("backup.fileExtension"),
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

	dumper := internal.NewDumper(backupConfig, mysqlDumpConfig, mysqlConfig)
	err = dumper.DumpDb()
	if err != nil {
		fmt.Printf("error in dumper: %s", err)
	}

	gzipper := internal.NewGzipper()
	err = gzipper.GzipFile(dumper.FileName)
	if err != nil {
		fmt.Printf("error while gzip file: %s", err)
	}

	deleter := internal.NewDeleter()
	err = deleter.DeleteFile(dumper.FileName)
	if err != nil {
		fmt.Printf("error while delete file: %s", err)
	}

	s3client := internal.NewS3Client(s3config)
	err = s3client.UploadFile(gzipper.FileName, true)
	if err != nil {
		fmt.Printf("error while uploading file to S3: %s", err)
	}

	err = deleter.DeleteFile(gzipper.FileName)
	if err != nil {
		fmt.Printf("error while deleting gz file: %s", err)
	}

	presignLink, err := s3client.GetPresignLink(gzipper.FileName, 24*time.Hour)
	if err != nil {
		fmt.Printf("error while generating presign link: %s", err)
	}

	tgClient := internal.NewTgClient(telegramConfig)
	tgClient.SendDbBackupFinishMessage(mysqlConfig.Name, presignLink)
}
