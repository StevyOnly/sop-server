package database

// 本文件是数据库连接的唯一收口选择器。
//
// 强制约定（代码评审清单）：
//  1. service 层一律只允许引用 Onebe() / SOPDB() / Sparepart()，
//     禁止直接引用 global.SOPhubDB / global.OnebeDB / global.SparepartDB 原始连接；
//  2. Onebe / SPAREPART 为只读库，代码层约定禁止一切写操作（Create/Update/Delete/Raw Exec），
//     硬性保障由数据库层最小权限只读账号提供（写操作会被数据库直接拒绝）。

import (
	"gorm.io/gorm"

	"sop/global"
)

// Onebe 老库只读查询。
// 代码层不设拦截（GORM DryRun 会连 SELECT 都不执行，不可用），
// 只读由数据库只读账号硬性保障；上线前必须将 Onebe 账号替换为只读账号。
func Onebe() *gorm.DB { return global.OnebeDB }

// Sparepart SPAREPART 备件老库只读查询。
// 与 Onebe 相同，只读由数据库只读账号硬性保障；上线前必须将账号替换为只读账号。
func Sparepart() *gorm.DB { return global.SparepartDB }

// SOPDB SOP 主库（SOPHub）读写
func SOPDB() *gorm.DB { return global.SOPhubDB }

// Close 关闭三个库的底层连接池（database/sql），服务优雅关闭时调用，
// 归还连接句柄，避免进程退出前连接悬留在数据库侧。
func Close() error {
	var firstErr error
	closeOne := func(db *gorm.DB) {
		if db == nil {
			return
		}
		sqlDB, err := db.DB()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			return
		}
		if err := sqlDB.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	closeOne(global.SOPhubDB)
	closeOne(global.OnebeDB)
	closeOne(global.SparepartDB)
	return firstErr
}
