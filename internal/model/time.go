package model

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// TimeLayout 全局统一时间格式：YYYY-MM-DD HH:mm:ss
const TimeLayout = "2006-01-02 15:04:05"

// LocalTime 统一时间类型，JSON 序列化输出 "YYYY-MM-DD HH:mm:ss"。
// 底层为 time.Time，可直接调用 time.Time 的方法（如 IsZero/Format）。
// 实现 driver.Valuer + sql.Scanner，保证 GORM 与 SQL Server 读写兼容，
// 以及 BaseModel 的 autoCreateTime/autoUpdateTime 正常生效。
type LocalTime time.Time

// Time 返回底层 time.Time
func (t LocalTime) Time() time.Time { return time.Time(t) }

// IsZero 判断是否为零值时间（空）
func (t LocalTime) IsZero() bool { return time.Time(t).IsZero() }

// Format 按指定布局格式化（如 TimeLayout）
func (t LocalTime) Format(layout string) string { return time.Time(t).Format(layout) }

// MarshalJSON 输出 "YYYY-MM-DD HH:mm:ss"
func (t LocalTime) MarshalJSON() ([]byte, error) {
	tt := time.Time(t)
	if tt.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(fmt.Sprintf("%q", tt.Format(TimeLayout))), nil
}

// UnmarshalJSON 兼容解析：
//   - "2006-01-02 15:04:05"（本项目统一格式）
//   - RFC3339（如旧数据 "2024-01-01T00:00:00+08:00"）
//   - 空字符串 / null
func (t *LocalTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*t = LocalTime(time.Time{})
		return nil
	}
	for _, layout := range []string{TimeLayout, time.RFC3339, "2006-01-02"} {
		if tt, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			*t = LocalTime(tt)
			return nil
		}
	}
	return fmt.Errorf("无法解析时间 %q", s)
}

// Value 实现 driver.Valuer
func (t LocalTime) Value() (driver.Value, error) {
	return time.Time(t), nil
}

// Scan 实现 sql.Scanner
func (t *LocalTime) Scan(v interface{}) error {
	if v == nil {
		*t = LocalTime(time.Time{})
		return nil
	}
	switch val := v.(type) {
	case time.Time:
		*t = LocalTime(val)
	case *time.Time:
		if val == nil {
			*t = LocalTime(time.Time{})
		} else {
			*t = LocalTime(*val)
		}
	case string:
		tt, err := time.ParseInLocation(TimeLayout, val, time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(tt)
	case []byte:
		tt, err := time.ParseInLocation(TimeLayout, string(val), time.Local)
		if err != nil {
			return err
		}
		*t = LocalTime(tt)
	default:
		return fmt.Errorf("无法扫描类型 %T 为 LocalTime", v)
	}
	return nil
}
