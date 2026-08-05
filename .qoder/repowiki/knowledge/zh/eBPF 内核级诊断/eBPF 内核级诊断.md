---
kind: external_dependency
name: eBPF 内核级诊断
slug: ebpf-cilium
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### 内核级性能与安全分析
- **角色**：利用 eBPF 技术进行无侵入式的系统级监控和诊断
- **集成点**：`internal/diag/ebpf/` 包含 analyzer、bpf、probe 三个子模块
- **技术栈**：基于 Cilium 的 ebpf 库实现
- **功能范围**：网络、进程、系统调用级别的深度诊断
- **部署要求**：需要内核支持 eBPF，通常在较新版本的 Linux 内核上
- **权限要求**：需要 CAP_BPF、CAP_SYS_ADMIN 等高级权限