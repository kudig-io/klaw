package network

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kudig-io/klaw/internal/diag/types"
)

func TestNewInterfaceAnalyzer(t *testing.T) {
	a := NewInterfaceAnalyzer()
	if a == nil {
		t.Fatal("NewInterfaceAnalyzer() returned nil")
	}
	if a.Name() != "network.interface" {
		t.Errorf("Name() = %v, want %v", a.Name(), "network.interface")
	}
}

func TestInterfaceAnalyzerAnalyze_NoFile(t *testing.T) {
	a := NewInterfaceAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestInterfaceAnalyzerAnalyze_NoIssue(t *testing.T) {
	a := NewInterfaceAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("eth0: <UP,BROADCAST> mtu 1500 state UP"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestInterfaceAnalyzerAnalyze_WithDownInterface(t *testing.T) {
	a := NewInterfaceAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("eth1: <BROADCAST> mtu 1500 state DOWN"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NETWORK_INTERFACE_DOWN" {
		t.Errorf("ENName = %v, want %v", issues[0].ENName, "NETWORK_INTERFACE_DOWN")
	}
}

func TestInterfaceAnalyzerAnalyze_ExcludeVeth(t *testing.T) {
	a := NewInterfaceAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// veth interfaces should be excluded
	data.RawFiles = map[string][]byte{
		"network_info": []byte("veth123: <BROADCAST> mtu 1500 state DOWN"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// veth interfaces should be excluded
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for veth, got %d", len(issues))
	}
}

func TestInterfaceAnalyzerAnalyze_ExcludeLo(t *testing.T) {
	a := NewInterfaceAnalyzer()
	ctx := context.Background()
	data := types.NewDiagnosticData(types.ModeOffline)
	// lo interface should be excluded
	data.RawFiles = map[string][]byte{
		"network_info": []byte("lo: <BROADCAST> mtu 1500 state DOWN"),
	}

	issues, err := a.Analyze(ctx, data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	// lo should be excluded
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for lo, got %d", len(issues))
	}
}

func TestNewRouteAnalyzer(t *testing.T) {
	a := NewRouteAnalyzer()
	if a == nil {
		t.Fatal("NewRouteAnalyzer() returned nil")
	}
	if a.Name() != "network.route" {
		t.Errorf("Name() = %v, want %v", a.Name(), "network.route")
	}
}

func TestNewIptablesAnalyzer(t *testing.T) {
	a := NewIptablesAnalyzer()
	if a == nil {
		t.Fatal("NewIptablesAnalyzer() returned nil")
	}
	if a.Name() != "network.iptables" {
		t.Errorf("Name() = %v, want %v", a.Name(), "network.iptables")
	}
}

func TestNewInodeAnalyzer(t *testing.T) {
	a := NewInodeAnalyzer()
	if a == nil {
		t.Fatal("NewInodeAnalyzer() returned nil")
	}
	if a.Name() != "network.inode" {
		t.Errorf("Name() = %v, want %v", a.Name(), "network.inode")
	}
}

func TestRouteAnalyzer_Analyze_NoFile(t *testing.T) {
	a := NewRouteAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestRouteAnalyzer_Analyze_NoDefaultRoute(t *testing.T) {
	a := NewRouteAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("10.0.0.0/24 via 10.0.0.1 dev eth0"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "NO_DEFAULT_ROUTE" {
		t.Errorf("ENName = %v, want NO_DEFAULT_ROUTE", issues[0].ENName)
	}
}

func TestRouteAnalyzer_Analyze_HasDefaultRoute(t *testing.T) {
	a := NewRouteAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("default via 10.0.0.1 dev eth0\n10.0.0.0/24 dev eth0"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestNewPortAnalyzer(t *testing.T) {
	a := NewPortAnalyzer()
	if a == nil {
		t.Fatal("NewPortAnalyzer() returned nil")
	}
	if a.Name() != "network.port" {
		t.Errorf("Name() = %v, want network.port", a.Name())
	}
}

func TestPortAnalyzer_Analyze_NoFile(t *testing.T) {
	a := NewPortAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestPortAnalyzer_Analyze_KubeletListening(t *testing.T) {
	a := NewPortAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("tcp  0  0 0.0.0.0:10250  0.0.0.0:*  LISTEN"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues when kubelet is listening, got %d", len(issues))
	}
}

func TestPortAnalyzer_Analyze_KubeletNotListening(t *testing.T) {
	a := NewPortAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("tcp  0  0 0.0.0.0:80  0.0.0.0:*  LISTEN"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "KUBELET_PORT_NOT_LISTENING" {
		t.Errorf("ENName = %v, want KUBELET_PORT_NOT_LISTENING", issues[0].ENName)
	}
}

func TestIptablesAnalyzer_Analyze_NoFile(t *testing.T) {
	a := NewIptablesAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestIptablesAnalyzer_Analyze_NormalRules(t *testing.T) {
	a := NewIptablesAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("-A INPUT -s 10.0.0.0/8 -j ACCEPT\n-A INPUT -j DROP"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal rule count, got %d", len(issues))
	}
}

func TestIptablesAnalyzer_Analyze_TooManyRules(t *testing.T) {
	a := NewIptablesAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	var rules string
	for i := 0; i < 50001; i++ {
		rules += fmt.Sprintf("\n-A INPUT -s 10.0.%d.%d -j ACCEPT", i/256, i%256)
	}
	data.RawFiles = map[string][]byte{
		"network_info": []byte(rules),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "TOO_MANY_IPTABLES_RULES" {
		t.Errorf("ENName = %v, want TOO_MANY_IPTABLES_RULES", issues[0].ENName)
	}
}

func TestInodeAnalyzer_Analyze_NoFile(t *testing.T) {
	a := NewInodeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

func TestInodeAnalyzer_Analyze_HighUsage(t *testing.T) {
	a := NewInodeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("/dev/sda1  1000000  950000  50000  95%  /"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}
	if issues[0].ENName != "HIGH_INODE_USAGE" {
		t.Errorf("ENName = %v, want HIGH_INODE_USAGE", issues[0].ENName)
	}
}

func TestInodeAnalyzer_Analyze_NormalUsage(t *testing.T) {
	a := NewInodeAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"system_status": []byte("/dev/sda1  1000000  500000  500000  50%  /"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues for normal usage, got %d", len(issues))
	}
}

func TestInterfaceAnalyzerAnalyze_MultipleDownInterfaces(t *testing.T) {
	a := NewInterfaceAnalyzer()
	data := types.NewDiagnosticData(types.ModeOffline)
	data.RawFiles = map[string][]byte{
		"network_info": []byte("eth1: <BROADCAST> mtu 1500 state DOWN\neth2: <BROADCAST> mtu 1500 state DOWN"),
	}
	issues, err := a.Analyze(context.Background(), data)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("Expected 1 combined issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Details, "eth1") || !strings.Contains(issues[0].Details, "eth2") {
		t.Errorf("Expected both interfaces in details, got: %s", issues[0].Details)
	}
}
