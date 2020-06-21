package types

// MySql конфиг для подключения к MySQL базе
type MySql struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// NewMySql конструктор
func NewMySql(host, port, name, user, password string) *MySql {
	return &MySql{
		Host:     host,
		Port:     port,
		Name:     name,
		User:     user,
		Password: password,
	}
}
