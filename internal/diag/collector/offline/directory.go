// Package offline implements the offline data collector
package offline

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kudig-io/klaw/internal/diag/collector"
	"github.com/kudig-io/klaw/internal/diag/types"
)

// Collector reads diagnostic data from a directory collected by diagnose_k8s.sh
type Collector struct{}

// NewCollector creates a new offline collector
func NewCollector() *Collector {
	return &Collector{}
}

// Name returns the collector name
func (c *Collector) Name() string {
	return "offline"
}

// Mode returns the data collection mode
func (c *Collector) Mode() types.DataMode {
	return types.ModeOffline
}

// Validate checks if the collector can operate with given config
func (c *Collector) Validate(config *collector.Config) error {
	if config.DiagnosePath == "" {
		return fmt.Errorf("diagnose_path is required for offline mode")
	}

	info, err := os.Stat(config.DiagnosePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("diagnostic directory does not exist: %s", config.DiagnosePath)
		}
		return fmt.Errorf("failed to access diagnostic directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", config.DiagnosePath)
	}

	return nil
}

// Collect gathers diagnostic data from the directory
func (c *Collector) Collect(ctx context.Context, config *collector.Config) (*types.DiagnosticData, error) {
	if err := c.Validate(config); err != nil {
		return nil, err
	}

	data := types.NewDiagnosticData(types.ModeOffline)
	data.DiagnosePath = config.DiagnosePath

	// Read key files
	keyFiles := []string{
		"system_info",
		"system_status",
		"service_status",
		"memory_info",
		"network_info",
		"ps_command_status",
	}

	for _, file := range keyFiles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		filePath := filepath.Join(config.DiagnosePath, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("failed to read %s: %w", file, err)
			}
			// File doesn't exist, skip
			continue
		}
		data.SetRawFile(file, content)
	}

	// Read daemon_status directory
	daemonStatusDir := filepath.Join(config.DiagnosePath, "daemon_status")
	if info, err := os.Stat(daemonStatusDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(daemonStatusDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				filePath := filepath.Join(daemonStatusDir, entry.Name())
				content, err := os.ReadFile(filePath)
				if err == nil {
					data.SetRawFile("daemon_status/"+entry.Name(), content)
				}
			}
		}
	}

	// Read logs directory
	logsDir := filepath.Join(config.DiagnosePath, "logs")
	if info, err := os.Stat(logsDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(logsDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				filePath := filepath.Join(logsDir, entry.Name())
				content, err := os.ReadFile(filePath)
				if err == nil {
					data.SetRawFile("logs/"+entry.Name(), content)
				}
			}
		}
	}

	// Parse node info
	data.NodeInfo = c.parseNodeInfo(data)

	// Parse system metrics
	data.SystemMetrics = c.parseSystemMetrics(data)

	return data, nil
}

// parseNodeInfo extracts node information from raw files
func (c *Collector) parseNodeInfo(data *types.DiagnosticData) types.NodeInfo {
	info := types.NodeInfo{}

	systemInfo, ok := data.GetRawFile("system_info")
	if !ok {
		return info
	}

	content := string(systemInfo)

	// Extract hostname
	hostnameRe := regexp.MustCompile(`(?i)hostname[:\s]+(\S+)`)
	if matches := hostnameRe.FindStringSubmatch(content); len(matches) > 1 {
		info.Hostname = matches[1]
	}

	// Extract kernel version from uname
	kernelRe := regexp.MustCompile(`Linux\s+\S+\s+(\S+)`)
	if matches := kernelRe.FindStringSubmatch(content); len(matches) > 1 {
		info.KernelVersion = matches[1]
	}

	// Extract OS image
	osRe := regexp.MustCompile(`(?i)PRETTY_NAME="([^"]+)"`)
	if matches := osRe.FindStringSubmatch(content); len(matches) > 1 {
		info.OSImage = matches[1]
	}

	// Extract kubelet version from daemon_status/kubelet_status
	kubeletStatus, ok := data.GetRawFile("daemon_status/kubelet_status")
	if ok {
		versionRe := regexp.MustCompile(`kubelet.*version[:\s]+v?(\S+)`)
		if matches := versionRe.FindStringSubmatch(string(kubeletStatus)); len(matches) > 1 {
			info.KubeletVersion = matches[1]
		}
	}

	return info
}

// parseSystemMetrics extracts system metrics from raw files
func (c *Collector) parseSystemMetrics(data *types.DiagnosticData) *types.SystemMetrics {
	metrics := &types.SystemMetrics{}

	// Parse CPU cores from system_info
	systemInfo, ok := data.GetRawFile("system_info")
	if ok {
		content := string(systemInfo)
		// Count processor entries
		processorCount := strings.Count(content, "processor\t:")
		if processorCount > 0 {
			metrics.CPUCores = processorCount
		} else {
			// Try CPU(s): pattern
			cpuRe := regexp.MustCompile(`CPU\(s\):\s*(\d+)`)
			if matches := cpuRe.FindStringSubmatch(content); len(matches) > 1 {
				if cores, err := strconv.Atoi(matches[1]); err == nil {
					metrics.CPUCores = cores
				}
			}
		}

		// Default to 4 if not found
		if metrics.CPUCores == 0 {
			metrics.CPUCores = 4
		}
	}

	// Parse load average from system_status
	systemStatus, ok := data.GetRawFile("system_status")
	if ok {
		content := string(systemStatus)
		loadRe := regexp.MustCompile(`load average:\s*([\d.]+),\s*([\d.]+),\s*([\d.]+)`)
		if matches := loadRe.FindStringSubmatch(content); len(matches) > 3 {
			metrics.LoadAvg1Min, _ = strconv.ParseFloat(matches[1], 64)
			metrics.LoadAvg5Min, _ = strconv.ParseFloat(matches[2], 64)
			metrics.LoadAvg15Min, _ = strconv.ParseFloat(matches[3], 64)
		}
	}

	// Parse memory from memory_info
	memoryInfo, ok := data.GetRawFile("memory_info")
	if ok {
		content := string(memoryInfo)

		memTotalRe := regexp.MustCompile(`MemTotal:\s*(\d+)`)
		if matches := memTotalRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.MemTotal, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		memAvailRe := regexp.MustCompile(`MemAvailable:\s*(\d+)`)
		if matches := memAvailRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.MemAvailable, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		memFreeRe := regexp.MustCompile(`MemFree:\s*(\d+)`)
		if matches := memFreeRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.MemFree, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		swapTotalRe := regexp.MustCompile(`SwapTotal:\s*(\d+)`)
		if matches := swapTotalRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.SwapTotal, _ = strconv.ParseInt(matches[1], 10, 64)
		}

		swapFreeRe := regexp.MustCompile(`SwapFree:\s*(\d+)`)
		if matches := swapFreeRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.SwapFree, _ = strconv.ParseInt(matches[1], 10, 64)
		}
	}

	// Parse disk usage from system_status
	if systemStatus, ok := data.GetRawFile("system_status"); ok {
		metrics.DiskUsage = c.parseDiskUsage(string(systemStatus))
	}

	// Parse conntrack from network_info
	if networkInfo, ok := data.GetRawFile("network_info"); ok {
		content := string(networkInfo)
		// Count current connections
		metrics.ConntrackCurrent = int64(strings.Count(content, "ipv4"))
	}

	// Get conntrack max from system_info
	if systemInfo, ok := data.GetRawFile("system_info"); ok {
		content := string(systemInfo)
		conntrackMaxRe := regexp.MustCompile(`net\.netfilter\.nf_conntrack_max\s*=\s*(\d+)`)
		if matches := conntrackMaxRe.FindStringSubmatch(content); len(matches) > 1 {
			metrics.ConntrackMax, _ = strconv.ParseInt(matches[1], 10, 64)
		}
	}

	return metrics
}

// parseDiskUsage extracts disk usage from df output
func (c *Collector) parseDiskUsage(content string) []types.DiskUsage {
	var result []types.DiskUsage

	// Find df -h section
	scanner := bufio.NewScanner(strings.NewReader(content))
	inDfSection := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "run df -h") || strings.Contains(line, "run df") {
			inDfSection = true
			continue
		}
		if inDfSection && strings.Contains(line, "End of df") {
			break
		}

		if inDfSection && strings.HasPrefix(line, "/") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				usage := types.DiskUsage{
					Device:     fields[0],
					MountPoint: fields[len(fields)-1],
				}

				// Parse usage percentage
				usedPercStr := strings.TrimSuffix(fields[len(fields)-2], "%")
				if percent, err := strconv.ParseFloat(usedPercStr, 64); err == nil {
					usage.UsedPercent = percent
				}

				result = append(result, usage)
			}
		}
	}

	return result
}

// init registers the collector with the default factory
func init() {
	collector.RegisterCollector(NewCollector())
}
