# Klaw Favicon 设计

## 设计说明

为 Klaw (Kubernetes Management Tool) 设计了多个 favicon 版本。

## 版本对比

### 1. favicon.svg (最终版本 - 推荐使用)
- **风格**: 现代工具风格
- **设计**: 渐变蓝色背景 + 白色工具形状 K
- **特点**: 
  - 圆角正方形，柔和专业
  - 蓝色渐变 (#3b82f6 → #1d4ed8)
  - K 字母设计成钳子/工具形状
  - 小装饰圆点增加视觉层次

### 2. favicon-v2.svg (渐变字母版)
- **风格**: 简约现代
- **设计**: 蓝色渐变背景 + 白色立体 K
- **特点**:
  - 更简洁的字母设计
  - 双渐变效果
  - 适合小尺寸显示

### 3. favicon-v3.svg (K8s 舵轮版)
- **风格**: Kubernetes 原生风格
- **设计**: 深色背景 + K8s 舵轮 + K 字母
- **特点**:
  - 致敬 Kubernetes 设计
  - 舵轮元素 + K 字母叠加
  - 深色主题，适合暗色模式

## 技术规格

- **格式**: SVG (可缩放矢量图形)
- **尺寸**: 100x100 viewBox
- **圆角**: 18-24px
- **色彩**: 蓝色系，符合 Kubernetes 生态

## 使用

当前使用的是 `favicon.svg` (最终版本)。

如需切换版本，修改 `index.html`:
```html
<link rel="icon" type="image/svg+xml" href="/favicon-v2.svg" />
```

## 浏览器兼容性

SVG favicon 支持:
- ✅ Chrome/Edge (推荐)
- ✅ Firefox
- ✅ Safari 12+
- ⚠️ IE11 不支持 (显示默认图标)
