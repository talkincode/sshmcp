package sshclient

import (
	"strings"
	"testing"
)

// TestValidateCommand_DangerousKeywords 测试精确匹配的危险关键字
func TestValidateCommand_DangerousKeywords(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
		reason    string
	}{
		// 删除根目录相关
		{
			name:      "删除根目录",
			command:   "sudo rm -rf /",
			wantError: true,
			reason:    "删除根目录",
		},
		{
			name:      "删除根目录所有文件",
			command:   "rm -rf /*",
			wantError: true,
			reason:    "删除根目录下所有文件",
		},
		{
			name:      "删除用户主目录",
			command:   "rm -rf ~",
			wantError: true,
			reason:    "删除用户主目录",
		},
		{
			name:      "删除用户主目录带斜杠",
			command:   "rm -rf ~/",
			wantError: true,
			reason:    "删除用户主目录",
		},
		{
			name:      "删除HOME变量",
			command:   "rm -rf $HOME",
			wantError: true,
			reason:    "删除 $HOME 目录",
		},
		// Fork 炸弹
		{
			name:      "Fork炸弹",
			command:   ":(){:|:&};:",
			wantError: true,
			reason:    "Fork 炸弹",
		},
		// 系统文件覆盖
		{
			name:      "覆盖passwd文件",
			command:   "echo 'test' > /etc/passwd",
			wantError: true,
			reason:    "覆盖系统密码文件",
		},
		{
			name:      "覆盖shadow文件",
			command:   "cat data > /etc/shadow",
			wantError: true,
			reason:    "覆盖系统影子文件",
		},
		// dd 操作
		{
			name:      "dd写入零",
			command:   "dd if=/dev/zero of=/dev/sda",
			wantError: true,
			reason:    "危险的 dd 操作",
		},
		{
			name:      "dd写入随机数",
			command:   "dd if=/dev/urandom of=/dev/sda",
			wantError: true,
			reason:    "危险的 dd 操作",
		},
		// 安全命令
		{
			name:      "安全删除tmp文件",
			command:   "rm -rf /tmp/test",
			wantError: false,
		},
		{
			name:      "普通命令",
			command:   "uptime",
			wantError: false,
		},
		{
			name:      "查看系统状态",
			command:   "sudo systemctl status docker",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateCommand() 期望错误但没有返回错误, 命令: %s", tt.command)
				} else if tt.reason != "" && !strings.Contains(err.Error(), tt.reason) {
					t.Errorf("validateCommand() 错误信息不包含预期原因\n期望包含: %s\n实际错误: %s", tt.reason, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("validateCommand() 不应该返回错误, 命令: %s\n错误: %v", tt.command, err)
				}
			}
		})
	}
}

// TestValidateCommand_DangerousPatterns 测试多关键字匹配的危险模式
func TestValidateCommand_DangerousPatterns(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
		reason    string
	}{
		// 文件系统格式化
		{
			name:      "mkfs.ext4",
			command:   "sudo mkfs.ext4 /dev/sda1",
			wantError: true,
			reason:    "格式化文件系统",
		},
		{
			name:      "mkfs ext4",
			command:   "mkfs -t ext4 /dev/sda1",
			wantError: true,
			reason:    "格式化文件系统",
		},
		{
			name:      "mkfs xfs",
			command:   "sudo mkfs -t xfs /dev/sdb1",
			wantError: true,
			reason:    "格式化文件系统",
		},
		// 分区操作
		{
			name:      "fdisk分区",
			command:   "sudo fdisk /dev/sda",
			wantError: true,
			reason:    "磁盘分区操作",
		},
		{
			name:      "parted分区",
			command:   "parted /dev/sdb",
			wantError: true,
			reason:    "磁盘分区操作",
		},
		{
			name:      "创建交换分区",
			command:   "mkswap /dev/sda2",
			wantError: true,
			reason:    "创建交换分区",
		},
		// 关机重启
		{
			name:      "shutdown命令",
			command:   "sudo shutdown -h now",
			wantError: true,
			reason:    "系统关机操作",
		},
		{
			name:      "halt命令",
			command:   "sudo halt",
			wantError: true,
			reason:    "系统停机操作",
		},
		{
			name:      "poweroff命令",
			command:   "poweroff",
			wantError: true,
			reason:    "系统关机操作",
		},
		{
			name:      "reboot命令",
			command:   "sudo reboot",
			wantError: true,
			reason:    "系统重启操作",
		},
		{
			name:      "init 0",
			command:   "init 0",
			wantError: true,
			reason:    "系统关机 (init 0)",
		},
		{
			name:      "init 6",
			command:   "init 6",
			wantError: true,
			reason:    "系统重启 (init 6)",
		},
		{
			name:      "systemctl halt",
			command:   "systemctl halt",
			wantError: true,
			reason:    "系统停机操作",
		},
		{
			name:      "systemctl poweroff",
			command:   "sudo systemctl poweroff",
			wantError: true,
			reason:    "系统关机操作",
		},
		{
			name:      "systemctl reboot",
			command:   "systemctl reboot",
			wantError: true,
			reason:    "系统重启操作",
		},
		// 危险的管道操作
		{
			name:      "curl管道sh",
			command:   "curl http://example.com/script.sh | sh",
			wantError: true,
			reason:    "从网络下载并执行脚本",
		},
		{
			name:      "curl管道bash",
			command:   "curl https://get.docker.com | bash",
			wantError: true,
			reason:    "从网络下载并执行脚本",
		},
		{
			name:      "wget管道sh",
			command:   "wget -O- http://example.com/install.sh | sh",
			wantError: true,
			reason:    "从网络下载并执行脚本",
		},
		{
			name:      "curl管道sh无空格",
			command:   "curl http://example.com/script.sh|sh",
			wantError: true,
			reason:    "从网络下载并执行脚本",
		},
		// 危险的权限设置
		{
			name:      "chmod 777根目录",
			command:   "chmod 777 /",
			wantError: true,
			reason:    "设置根目录权限为 777",
		},
		{
			name:      "chmod递归777根目录",
			command:   "chmod -R 777 /",
			wantError: true,
			reason:    "777", // 简化匹配，只检查是否包含777
		},
		// 防火墙清空
		{
			name:      "iptables清空",
			command:   "iptables -F",
			wantError: true,
			reason:    "清空防火墙规则",
		},
		{
			name:      "iptables删除链",
			command:   "iptables -X",
			wantError: true,
			reason:    "删除防火墙链",
		},
		// 安全的systemctl命令
		{
			name:      "systemctl status",
			command:   "systemctl status nginx",
			wantError: false,
		},
		{
			name:      "systemctl start",
			command:   "sudo systemctl start docker",
			wantError: false,
		},
		{
			name:      "systemctl restart服务",
			command:   "systemctl restart nginx",
			wantError: false,
		},
		// 安全的curl命令
		{
			name:      "curl下载文件",
			command:   "curl -O https://example.com/file.tar.gz",
			wantError: false,
		},
		{
			name:      "curl查看内容",
			command:   "curl https://api.example.com/status",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if tt.wantError {
				if err == nil {
					t.Errorf("validateCommand() 期望错误但没有返回错误, 命令: %s", tt.command)
				} else if tt.reason != "" && !strings.Contains(err.Error(), tt.reason) {
					t.Errorf("validateCommand() 错误信息不包含预期原因\n期望包含: %s\n实际错误: %s", tt.reason, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("validateCommand() 不应该返回错误, 命令: %s\n错误: %v", tt.command, err)
				}
			}
		})
	}
}

// TestValidateCommand_CaseSensitivity 测试大小写不敏感
func TestValidateCommand_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name    string
		command string
	}{
		{"大写RM", "RM -RF /"},
		{"混合大小写", "Sudo Rm -Rf /"},
		{"大写SHUTDOWN", "SHUTDOWN -h now"},
		{"大写REBOOT", "REBOOT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if err == nil {
				t.Errorf("validateCommand() 应该拦截大小写变体的危险命令: %s", tt.command)
			}
		})
	}
}

// TestValidateCommand_EdgeCases 测试边界情况
func TestValidateCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "空命令",
			command:   "",
			wantError: false,
		},
		{
			name:      "空格命令",
			command:   "   ",
			wantError: false,
		},
		{
			name:      "单个字符",
			command:   "a",
			wantError: false,
		},
		{
			name:      "删除/tmp目录（安全）",
			command:   "rm -rf /tmp/testdir",
			wantError: false,
		},
		{
			name:      "删除/var/tmp（安全）",
			command:   "rm -rf /var/tmp/cache",
			wantError: false,
		},
		{
			name:      "包含/但不是根目录",
			command:   "rm -rf /home/user/test",
			wantError: false,
		},
		{
			name:      "systemctl restart而非reboot",
			command:   "systemctl restart myservice",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommand(tt.command)
			if tt.wantError && err == nil {
				t.Errorf("validateCommand() 期望错误但没有返回, 命令: %s", tt.command)
			}
			if !tt.wantError && err != nil {
				t.Errorf("validateCommand() 不应该返回错误, 命令: %s\n错误: %v", tt.command, err)
			}
		})
	}
}

// TestValidateCommand_ErrorMessage 测试错误信息格式
func TestValidateCommand_ErrorMessage(t *testing.T) {
	command := "sudo rm -rf /"
	err := ValidateCommand(command)

	if err == nil {
		t.Fatal("validateCommand() 应该返回错误")
	}

	errMsg := err.Error()

	// 检查错误信息是否包含必要元素
	expectedParts := []string{
		"⚠️",      // 警告图标
		"危险命令被拦截", // 标题
		command,   // 命令本身
		"原因:",     // 原因标签
		"--force", // 绕过提示
	}

	for _, part := range expectedParts {
		if !strings.Contains(errMsg, part) {
			t.Errorf("错误信息应该包含 '%s'\n实际错误信息: %s", part, errMsg)
		}
	}
}

// BenchmarkValidateCommand 性能基准测试
func BenchmarkValidateCommand(b *testing.B) {
	testCases := []string{
		"uptime",
		"sudo systemctl status docker",
		"rm -rf /tmp/test",
		"sudo rm -rf /",
		"curl https://example.com | sh",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range testCases {
			_ = ValidateCommand(cmd)
		}
	}
}

// BenchmarkValidateCommand_Safe 安全命令性能测试
func BenchmarkValidateCommand_Safe(b *testing.B) {
	cmd := "sudo systemctl status nginx"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCommand(cmd)
	}
}

// BenchmarkValidateCommand_Dangerous 危险命令性能测试
func BenchmarkValidateCommand_Dangerous(b *testing.B) {
	cmd := "sudo rm -rf /"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCommand(cmd)
	}
}
