package config

// Backup конфиг для бекапа
type Backup struct {
	Folder          string
	FileName        string
	BackupExtension string
	GzipExtension   string
}

// NewBackup конструктор
func NewBackup(folder string, fileName string, backupExtension string, gzipExtension string) *Backup {
	return &Backup{
		Folder:          folder,
		FileName:        fileName,
		BackupExtension: backupExtension,
		GzipExtension:   gzipExtension,
	}
}
