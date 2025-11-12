package sshclient

import (
	"fmt"
	"log"
	"strings"

	"github.com/zalando/go-keyring"
)

const KeyringServiceName = "sshx"

// ValidateCommand 验证命令安全性
func ValidateCommand(command string) error {
	cmd := strings.TrimSpace(command)
	cmdLower := strings.ToLower(cmd)

	dangerousExactPatterns := []struct {
		pattern string
		reason  string
	}{
		{" rm -rf / ", "删除根目录"},
		{" rm -rf /$", "删除根目录"},
		{" rm -rf /;", "删除根目录"},
		{" rm -rf /&", "删除根目录"},
		{" rm -rf /|", "删除根目录"},
		{"rm -rf / ", "删除根目录"},
		{"rm -rf /$", "删除根目录"},
		{"rm -rf /;", "删除根目录"},
		{"rm -rf /*", "删除根目录下所有文件"},
		{"rm -rf ~", "删除用户主目录"},
		{"rm -rf ~/", "删除用户主目录"},
		{"rm -rf $home", "删除 $HOME 目录"},
		{":(){:|:&};:", "Fork 炸弹"},
		{"> /etc/passwd", "覆盖系统密码文件"},
		{"> /etc/shadow", "覆盖系统影子文件"},
		{"dd if=/dev/zero", "危险的 dd 操作"},
		{"dd if=/dev/urandom", "危险的 dd 操作"},
	}

	for _, pattern := range dangerousExactPatterns {
		cmdWithSpaces := " " + cmdLower + " "
		patternLower := strings.ToLower(pattern.pattern)

		if strings.HasSuffix(pattern.pattern, "$") {
			patternLower = strings.TrimSuffix(patternLower, "$")
			if strings.HasSuffix(cmdLower, patternLower) {
				return fmt.Errorf("⚠️  危险命令被拦截\n命令: %s\n原因: %s\n如果确定要执行，请使用 --force 或 -f 参数", cmd, pattern.reason)
			}
		} else if strings.Contains(cmdWithSpaces, patternLower) {
			return fmt.Errorf("⚠️  危险命令被拦截\n命令: %s\n原因: %s\n如果确定要执行，请使用 --force 或 -f 参数", cmd, pattern.reason)
		}
	}

	dangerousPatterns := []struct {
		keywords []string
		reason   string
	}{
		{[]string{"mkfs."}, "格式化文件系统"},
		{[]string{"mkfs", "ext4"}, "格式化文件系统"},
		{[]string{"mkfs", "ext3"}, "格式化文件系统"},
		{[]string{"mkfs", "xfs"}, "格式化文件系统"},
		{[]string{"fdisk", "/dev/"}, "磁盘分区操作"},
		{[]string{"parted", "/dev/"}, "磁盘分区操作"},
		{[]string{"mkswap", "/dev/"}, "创建交换分区"},
		{[]string{"shutdown"}, "系统关机操作"},
		{[]string{"halt"}, "系统停机操作"},
		{[]string{"poweroff"}, "系统关机操作"},
		{[]string{"reboot"}, "系统重启操作"},
		{[]string{"init 0"}, "系统关机 (init 0)"},
		{[]string{"init 6"}, "系统重启 (init 6)"},
		{[]string{"systemctl", "halt"}, "系统停机操作"},
		{[]string{"systemctl", "poweroff"}, "系统关机操作"},
		{[]string{"systemctl", "reboot"}, "系统重启操作"},
		{[]string{"curl", "| sh"}, "从网络下载并执行脚本"},
		{[]string{"curl", "| bash"}, "从网络下载并执行脚本"},
		{[]string{"curl", "|sh"}, "从网络下载并执行脚本"},
		{[]string{"curl", "|bash"}, "从网络下载并执行脚本"},
		{[]string{"wget", "| sh"}, "从网络下载并执行脚本"},
		{[]string{"wget", "| bash"}, "从网络下载并执行脚本"},
		{[]string{"wget", "|sh"}, "从网络下载并执行脚本"},
		{[]string{"wget", "|bash"}, "从网络下载并执行脚本"},
		{[]string{"chmod", "777", "/ "}, "设置根目录权限为 777"},
		{[]string{"chmod", "777", "/$"}, "设置根目录权限为 777"},
		{[]string{"chmod", "-r", "777", "/ "}, "递归设置根目录权限为 777"},
		{[]string{"chmod", "-r", "777", "/$"}, "递归设置根目录权限为 777"},
		{[]string{"iptables", "-f"}, "清空防火墙规则"},
		{[]string{"iptables", "-x"}, "删除防火墙链"},
	}

	for _, pattern := range dangerousPatterns {
		allMatch := true
		for _, keyword := range pattern.keywords {
			keywordLower := strings.ToLower(keyword)
			if strings.HasSuffix(keyword, "$") {
				keywordLower = strings.TrimSuffix(keywordLower, "$")
				if !strings.HasSuffix(cmdLower, keywordLower) {
					allMatch = false
					break
				}
			} else if !strings.Contains(cmdLower, keywordLower) {
				allMatch = false
				break
			}
		}
		if allMatch {
			return fmt.Errorf("⚠️  危险命令被拦截\n命令: %s\n原因: %s\n如果确定要执行，请使用 --force 或 -f 参数", cmd, pattern.reason)
		}
	}

	return nil
}

// GetSudoPassword 从系统密钥环读取sudo密码（跨平台支持）
// macOS: Keychain, Linux: Secret Service (gnome-keyring/kwallet), Windows: Credential Manager
func GetSudoPassword(key string) (string, error) {
	serviceName := KeyringServiceName

	password, err := keyring.Get(serviceName, key)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", fmt.Errorf("sudo password not found in keyring for key: %s\n"+
				"Add it using one of:\n"+
				"  macOS:   security add-generic-password -s %s -a %s -w <password>\n"+
				"  Linux:   secret-tool store --label='Sudo Password' service %s username %s\n"+
				"  Windows: Use 'Credential Manager' in Control Panel",
				key, serviceName, key, serviceName, key)
		}
		return "", fmt.Errorf("failed to get sudo password from keyring: %w", err)
	}

	if password == "" {
		return "", fmt.Errorf("empty sudo password in keyring for key: %s", key)
	}

	log.Printf("✓ Sudo password loaded from system keyring for key: %s", key)
	return password, nil
}
