package initialize

import "fmt"

// InitConfig 读取配置文件并启动热重载监听
func InitConfig() error {
	if err := Init(); err != nil {
		return fmt.Errorf("init config failed: %w", err)
	}
	Watch()
	return nil
}
