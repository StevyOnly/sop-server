package config

// Config 全局配置
type Config struct {
	Server Server `mapstructure:"server"`
	// DatabaseOnebe 为 Onebe 老库（只读）配置
	DatabaseOnebe Database `mapstructure:"databaseOnebe"`
	// DatabaseSPAREPART 为 SPAREPART 备件老库（只读）配置
	DatabaseSPAREPART Database `mapstructure:"databaseSPAREPART"`
	// DatabaseSOPHub 为 SOPHub 新库（读写）配置
	DatabaseSOPHub Database  `mapstructure:"databaseSOPHub"`
	Logger         Logger    `mapstructure:"logger"`
	JWT            JWTConfig `mapstructure:"jwt"`
}

// JWTConfig JWT 相关配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpiresDays int    `mapstructure:"expiresDays"`
}

// Server 服务配置
type Server struct {
	// Mode gin 运行模式：debug / release / test（生产环境必须 release）
	Mode      string `mapstructure:"mode"`
	Port      int    `mapstructure:"port"`
	UploadDir string `mapstructure:"uploadDir"` // 上传文件存储根目录（静态资源映射根）
	// UploadMediaDir 媒体资源上传目录
	UploadMediaDir string `mapstructure:"uploadMediaDir"` // 上传文件存储目录：媒体资源
	// UploadSceneDir 现场照片上传目录
	UploadSceneDir string `mapstructure:"uploadSceneDir"` // 上传文件存储目录：现场照片
	// MaxScenePhotoSizeMB 现场照片 base64 上传大小上限（MB），默认 20
	MaxScenePhotoSizeMB int `mapstructure:"maxScenePhotoSizeMB"`
	// SkipPasswordCheck 登录是否跳过密码校验（测试用），true 时任意密码均可登录
	SkipPasswordCheck bool `mapstructure:"skipPasswordCheck"`
	// BaseURL 服务对外 URL 前缀（不含端口），如 http://192.168.100.10；用于拼装资源/分享链接的绝对地址（备用）
	BaseURL string `mapstructure:"baseUrl"`
	// StaticPort 静态资源对外端口（备用），与 server.port 相互独立时用于拼装静态资源完整访问地址
	StaticPort int `mapstructure:"staticPort"`
	// AppTokenKey APP 端 appLogin 换取永久 token 的固定密钥（"是否 APP 端"标志）。
	// 与 JWT secret 相互独立，是 APP 端令牌体系的唯一防线；生产必须改为强随机值，
	// 更换该值会导致所有已签发 APP 永久 token 全局失效（需 App 端重新 appLogin）。
	AppTokenKey string `mapstructure:"appTokenKey"`
}

// Database 数据库配置
type Database struct {
	// Driver 数据库驱动类型：mysql / sqlserver，缺省默认 sqlserver
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSL      bool   `mapstructure:"ssl"`

	// 连接池配置
	MaxOpenConns    int `mapstructure:"maxOpenConns"`
	MaxIdleConns    int `mapstructure:"maxIdleConns"`
	ConnMaxLifetime int `mapstructure:"connMaxLifetime"` // 秒
	ConnMaxIdleTime int `mapstructure:"connMaxIdleTime"` // 秒
}

// DriverType 返回数据库驱动类型，缺省为 sqlserver
func (d Database) DriverType() string {
	if d.Driver == "" {
		return "sqlserver"
	}
	return d.Driver
}

// SSLFlag 返回连接串中按驱动类型区分的加密参数
// SQLServer 连接默认开启加密，ssl=false 时禁用加密（encrypt=disable）；
// MySQL 通过 tls 参数控制，ssl=false 时禁用（tls=false）。
func (d Database) SSLFlag() string {
	if d.DriverType() == "mysql" {
		if d.SSL {
			return ""
		}
		return "&tls=false"
	}
	if d.SSL {
		return ""
	}
	return "&encrypt=disable"
}

// Logger 日志配置
type Logger struct {
	Level      string `mapstructure:"level"`
	Dir        string `mapstructure:"dir"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxAge     int    `mapstructure:"maxAge"`
	MaxBackups int    `mapstructure:"maxBackups"`
	Compress   bool   `mapstructure:"compress"`
	Cut        string `mapstructure:"cut"`
	Console    bool   `mapstructure:"console"` // 是否同时输出日志到终端控制台
}
