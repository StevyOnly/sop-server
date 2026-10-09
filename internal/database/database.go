// Package database 负责初始化 SQL Server 数据库连接（GORM）。
// 本系统使用三个数据库：
//   - SOPHub 新库：读写，全部 nxp_sop_* 业务表位于此库
//   - Onebe 老库：只读，仅供查询 nxp_user / nxp_facility / nxp_user_info，
//     只读由数据库层最小权限只读账号硬性保障（代码层约定禁止写，见 db.go）
//   - SPAREPART 老库：只读，仅供查询备件字典（sparepart）等数据，
//     只读由数据库层最小权限只读账号硬性保障（代码层约定禁止写，见 db.go）
package database

import (
	"fmt"
	"net/url"
	"time"

	"sop/global"
	"sop/internal/config"
	"sop/pkg/logger"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// InitSOPHub 初始化 SOPHub 新库（读写）连接
func InitSOPHub() error {
	db, err := InitOne(global.GetConfig().DatabaseSOPHub)
	if err != nil {
		return err
	}
	global.SOPhubDB = db
	global.Logger.Info("数据库连接成功: SOPHub (sqlserver)")
	return nil
}

// InitOnebe 初始化 Onebe 老库（只读）连接。
// 只读由数据库层最小权限只读账号保障（生产环境必须配置，当前沿用 sa 为已知遗留项）；
// 代码层禁止在此连接上执行任何写操作，约定见 db.go。
func InitOnebe() error {
	db, err := InitOne(global.GetConfig().DatabaseOnebe)
	if err != nil {
		return err
	}
	global.OnebeDB = db
	global.Logger.Info("数据库连接成功: Onebe (sqlserver, 只读)")
	return nil
}

// InitSparepart 初始化 SPAREPART 备件老库（只读）连接。
// 只读由数据库层最小权限只读账号保障（生产环境必须配置，当前沿用 sa 为已知遗留项）；
// 代码层禁止在此连接上执行任何写操作，约定见 db.go。
func InitSparepart() error {
	db, err := InitOne(global.GetConfig().DatabaseSPAREPART)
	if err != nil {
		return err
	}
	global.SparepartDB = db
	global.Logger.Info("数据库连接成功: SPAREPART (" + global.GetConfig().DatabaseSPAREPART.DriverType() + ", 只读)")
	return nil
}

// InitOne 根据配置建立数据库连接（支持 MySQL / SQL Server 两种驱动）
func InitOne(dbCfg config.Database) (*gorm.DB, error) {
	var (
		dsn    string
		dialec gorm.Dialector
	)

	switch dbCfg.DriverType() {
	case "mysql":
		// MySQL DSN 通过 mysql.Config + FormatDSN 生成：密码会被原样拼进 DSN，
		// 避免 url.URL 的 %xx 转义导致认证失配（go-sql-driver 解析时不会还原 %40/%21）
		dsnConf := mysqldriver.Config{
			User:   dbCfg.User,
			Passwd: dbCfg.Password,
			Net:    "tcp",
			Addr:   fmt.Sprintf("%s:%d", dbCfg.Host, dbCfg.Port),
			DBName: dbCfg.DBName,
			// ParseTime/Loc 与旧 DSN 保持一致：时间字段解析为 time.Time，时区用本地
			ParseTime: true,
			Loc:       time.Local,
			// MySQL 8 认证插件兼容：显式放开 mysql_native_password，
			// TLS 用 preferred（TLS 可用则用，否则允许回退明文，
			// 保证 caching_sha2/native 都能完成密码交换）
			AllowNativePasswords: true,
			TLSConfig:            "preferred",
		}
		dsn = dsnConf.FormatDSN()
		dialec = mysql.New(mysql.Config{DSNConfig: &dsnConf})
	default:
		// SQL Server DSN：sqlserver://user:pass@host:port?database=dbname
		dsn = fmt.Sprintf("sqlserver://%s@%s:%d?database=%s%s",
			// 用户名/密码做 URL 转义，避免口令含 # : @ / 等特殊字符时连接串被错误解析
			url.UserPassword(dbCfg.User, dbCfg.Password).String(),
			dbCfg.Host,
			dbCfg.Port,
			dbCfg.DBName,
			dbCfg.SSLFlag(),
		)
		dialec = sqlserver.Open(dsn)
	}

	db, err := gorm.Open(dialec, &gorm.Config{
		// SQL 日志接入 zap，等级为 Warn：只记录慢 SQL 与错误，便于与项目日志体系统一
		Logger: logger.NewGormLogger(global.Logger, logger.GormOptions{
			LogLevel: gormlogger.Warn,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库 %s 失败: %w", dbCfg.DBName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 连接池参数（未配置时使用 GORM/SQLServer 默认值）
	if dbCfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	}
	if cfg := dbCfg.MaxIdleConns; cfg > 0 {
		sqlDB.SetMaxIdleConns(cfg)
	}
	if cfg := dbCfg.ConnMaxLifetime; cfg > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg) * time.Second)
	}
	if cfg := dbCfg.ConnMaxIdleTime; cfg > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg) * time.Second)
	}
	return db, nil
}
