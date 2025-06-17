package config

type DBConfig struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func LoadAllDBConfigs() map[string]DBConfig {
	// bisa load dari env, file, atau hardcoded
	return map[string]DBConfig{
		"ifs": {
			Driver:   "mysql",
			Host:     "172.18.224.214",
			Port:     "3306",
			User:     "cbm_apps",
			Password: "P@ssw0rd",
			Name:     "new_ifs",
		},
	}
}
