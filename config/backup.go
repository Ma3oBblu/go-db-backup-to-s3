package config

// Backup конфиг для бекапа
type Backup struct {
	Folder    string
	Name      string
	Extension string
}

// NewBackup конструктор
func NewBackup(folder string, name string, extension string) *Backup {
	return &Backup{
		Folder:    folder,
		Name:      name,
		Extension: extension,
	}
}
