---
kind: dependency_management
name: 多模块依赖管理（Go Modules + npm Lockfile）
category: dependency_management
scope:
    - '**'
source_files:
    - go.mod
    - go.sum
    - operator/go.mod
    - operator/go.sum
    - modules/etcd-backup/go.mod
    - web/package.json
    - web/package-lock.json
---

本仓库采用多语言、多模块的依赖管理策略，通过 Go Modules 与 npm 锁文件分别管理后端与前端依赖，各子模块独立声明版本，未使用 vendor 目录或私有代理。