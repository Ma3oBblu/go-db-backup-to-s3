package types

// MySqlDump конфиг для mysqldump
type MySqlDump struct {
	IgnoreTable  string
	AddDropTable bool
}

// NewMySqlDump конструктор
func NewMySqlDump(ignoreTable string, addDropTable bool) *MySqlDump {
	return &MySqlDump{
		IgnoreTable:  ignoreTable,
		AddDropTable: addDropTable,
	}
}
