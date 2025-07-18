package model

type DBConf struct {
	Driver   string // 驱动
	Host     string // 主机地址
	Port     string // 主机端口
	User     string // 用户名
	Password string // 密码
	DBName   string // 数据库名称
	Charset  string // 编码
}

type RedisConf struct {
	Host     string // 主机地址
	Port     string // 主机端口
	Password string // 密码
}
