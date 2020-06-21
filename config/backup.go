package config

// Backup конфиг для бекапа
type Backup struct {
	Folder          string
	FileName        string
	BackupExtension string
}

// NewBackup конструктор
func NewBackup(folder, fileName, backupExtension string) *Backup {
	return &Backup{
		Folder:          folder,
		FileName:        fileName,
		BackupExtension: backupExtension,
	}
}
