#!/bin/bash
#
# Klaw Web 测试运行脚本
#

set -e

echo "🧪 Klaw Web Test Suite"
echo "======================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查是否在 web 目录
if [ ! -f "package.json" ]; then
    echo -e "${RED}Error: Please run this script from the web directory${NC}"
    exit 1
fi

# 显示帮助
show_help() {
    cat << EOF
Usage: ./test.sh [command]

Commands:
    all         运行所有测试
    unit        运行单元测试
    integration 运行集成测试
    coverage    运行测试并生成覆盖率报告
    ui          打开测试 UI
    watch       监听模式运行测试
    help        显示帮助

Examples:
    ./test.sh all           # 运行所有测试
    ./test.sh coverage      # 生成覆盖率报告
EOF
}

# 运行所有测试
run_all() {
    echo -e "${YELLOW}Running all tests...${NC}"
    npm run test:run
    echo -e "${GREEN}✅ All tests passed!${NC}"
}

# 运行单元测试
run_unit() {
    echo -e "${YELLOW}Running unit tests...${NC}"
    npm run test:run -- src/__tests__/unit
    echo -e "${GREEN}✅ Unit tests passed!${NC}"
}

# 运行集成测试
run_integration() {
    echo -e "${YELLOW}Running integration tests...${NC}"
    npm run test:run -- src/__tests__/integration
    echo -e "${GREEN}✅ Integration tests passed!${NC}"
}

# 生成覆盖率报告
run_coverage() {
    echo -e "${YELLOW}Running tests with coverage...${NC}"
    npm run test:coverage
    echo -e "${GREEN}✅ Coverage report generated!${NC}"
    echo ""
    echo "View report: web/coverage/index.html"
}

# 打开测试 UI
run_ui() {
    echo -e "${YELLOW}Opening test UI...${NC}"
    npm run test:ui
}

# 监听模式
run_watch() {
    echo -e "${YELLOW}Running tests in watch mode...${NC}"
    npm test
}

# 主命令处理
case "${1:-all}" in
    all)
        run_all
        ;;
    unit)
        run_unit
        ;;
    integration)
        run_integration
        ;;
    coverage)
        run_coverage
        ;;
    ui)
        run_ui
        ;;
    watch)
        run_watch
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        show_help
        exit 1
        ;;
esac
