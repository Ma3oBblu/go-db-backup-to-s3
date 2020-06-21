package config

// MySql конфиг для подключения к MySQL базе
type MySql struct {
	Name     string
	User     string
	Password string
}

// NewMySql конструктор
func NewMySql(name, user, password string) *MySql {
	return &MySql{
		Name:     name,
		User:     user,
		Password: password,
	}
}
