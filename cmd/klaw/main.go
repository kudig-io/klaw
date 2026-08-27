package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "klaw",
	Short: "Kubernetes 全场景运维平台 — 诊断 + 管理 + ChatOps",
	Long: `klaw — 融合了 kudig 诊断核心、k8s-guardian 运维管理和 ChatOps 的统一平台。

子命令:
  klaw server    启动 Web API + ChatOps 服务
  klaw diag      对集群运行诊断分析 (70+ 分析器)
  klaw version   打印版本信息`,
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Fprintln(cmd.OutOrStdout(), "klaw v1.0.0-fusion (诊断核心已集成)")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if len(os.Args) <= 1 {
		os.Args = []string{os.Args[0], "server"}
	} else {
		switch os.Args[1] {
		case "server", "diag", "version", "help", "-h", "--help", "completion":
		default:
			os.Args = append([]string{os.Args[0], "server"}, os.Args[1:]...)
		}
	}
	_ = rootCmd.Execute()
}
