package internal

import (
	"bufio"
	"fmt"
	"go-db-backup-to-s3/config"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

// Dumper дампер БД с помощью mysqldump
type Dumper struct {
	BackupConfig  *config.Backup
	SqlDumpConfig *config.MySqlDump
	SqlConfig     *config.MySql
	FileName      string
}

// NewDumper конструктор
func NewDumper(
	backupConfig *config.Backup,
	sqlDumpConfig *config.MySqlDump,
	sqlConfig *config.MySql,
) *Dumper {
	return &Dumper{
		BackupConfig:  backupConfig,
		SqlDumpConfig: sqlDumpConfig,
		SqlConfig:     sqlConfig,
	}
}

// generateBackupDate генерирует дату бекапа
func (d *Dumper) generateBackupDate() string {
	dt := time.Now()
	return strconv.Itoa(dt.Year()) + "-" + dt.Weekday().String()
}

// DumpDb дампит базу данных с помощью mysqldump в указанный файл backupFileFullPath
func (d *Dumper) DumpDb() error {
	backupFullPath := d.BackupConfig.Folder + d.BackupConfig.FileName + "." + d.generateBackupDate() + d.BackupConfig.BackupExtension

	mysqlDumpExtras := ""
	if d.SqlDumpConfig.IgnoreTable != "" {
		mysqlDumpExtras = "--ignore-table=" + d.SqlDumpConfig.IgnoreTable
	}
	if d.SqlDumpConfig.AddDropTable == true {
		mysqlDumpExtras += " --add-drop-table"
	}
	cmd := exec.Command(
		"mysqldump",
		"-u"+d.SqlConfig.User,
		"-p"+d.SqlConfig.Password,
		d.SqlConfig.Name,
		mysqlDumpExtras,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	outfile, err := os.Create(backupFullPath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer outfile.Close()

	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return err
	}

	// read command's stdout line by line
	in := bufio.NewWriter(outfile)
	defer in.Flush()

	if _, err := io.Copy(outfile, stdout); err != nil {
		log.Fatal(err)
		return err
	}

	d.FileName = backupFullPath
	fmt.Println("finish dump")
	return nil
}
