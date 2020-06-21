package config

// Backup конфиг для бекапа
type Backup struct {
	Folder          string
	BackupExtension string
}

// NewBackup конструктор
func NewBackup(folder, backupExtension string) *Backup {
	return &Backup{
		Folder:          folder,
		BackupExtension: backupExtension,
	}
}
