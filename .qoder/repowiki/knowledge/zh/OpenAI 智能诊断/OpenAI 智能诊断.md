---
kind: external_dependency
name: OpenAI 智能诊断
slug: openai
category: external_dependency
category_hints:
    - vendor_identity
scope:
    - '**'
---

### AI 驱动的诊断分析
- **角色**：为诊断结果提供智能分析和修复建议
- **集成点**：`internal/diag/ai/` 实现 OpenAI API 客户端和 provider 接口
- **功能特性**：自然语言分析、问题根因定位、自动修复建议
- **配置要求**：需要有效的 OpenAI API Key
- **扩展性**：provider 接口设计支持接入其他 AI 服务
- **成本考虑**：API 调用会产生费用，需注意使用量控制