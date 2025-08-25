# 🐳 Docker & CI/CD Configuration

本目錄包含 Repo Onboarding Copilot 的 Docker 配置和 CI/CD 自動化設定。

## 📁 檔案結構

```
.docker/
├── Dockerfile              # 生產環境 Docker 映像
├── Dockerfile.dev          # 開發環境 Docker 映像 (含熱重載)
├── docker-compose.yml      # Docker Compose 配置
├── docker-dev.sh           # 開發環境管理腳本
├── .env.example            # 環境變數範例
└── README.md               # 本文件
```

## 🚀 快速開始

### 1. 環境準備

```bash
# 複製環境變數檔案
cp .docker/.env.example .docker/.env

# 編輯環境變數（可選）
nano .docker/.env

# 使腳本可執行
chmod +x .docker/docker-dev.sh
```

### 2. 開發環境

```bash
# 啟動基本開發環境（僅應用程式）
.docker/docker-dev.sh start-dev

# 啟動完整開發環境（含資料庫、快取）
.docker/docker-dev.sh start-full

# 啟動監控環境（含 Prometheus、Grafana）
.docker/docker-dev.sh start-monitor

# 查看日誌
.docker/docker-dev.sh logs

# 進入開發容器
.docker/docker-dev.sh shell

# 停止所有服務
.docker/docker-dev.sh stop
```

### 3. 本地測試

```bash
# 在容器中執行測試
.docker/docker-dev.sh test

# 在容器中執行 linting
.docker/docker-dev.sh lint

# 在容器中執行安全掃描
.docker/docker-dev.sh security
```

## 🏗️ Docker 映像

### 生產環境映像特點
- **基於 Alpine Linux**: 最小化安全攻擊面
- **多階段建置**: 減少映像大小
- **非 root 使用者**: 增強安全性
- **健康檢查**: 自動監控應用狀態
- **安全配置**: Seccomp、AppArmor 支援

### 開發環境映像特點
- **熱重載**: 使用 Air 工具自動重新編譯
- **除錯支援**: Delve 除錯器整合
- **開發工具**: 包含 linting、安全掃描工具
- **Docker-in-Docker**: 支援容器分析功能

## 🔧 服務配置

### 核心服務

| 服務 | 端口 | 說明 |
|------|------|------|
| **app** | 8080 | 主應用程式 (生產) |
| **dev** | 8080, 2345 | 開發環境 (含除錯器) |

### 擴展服務 (Profiles)

| Profile | 服務 | 端口 | 說明 |
|---------|------|------|------|
| `cache` | redis | 6379 | Redis 快取 |
| `database` | postgres | 5432 | PostgreSQL 資料庫 |
| `monitoring` | prometheus | 9090 | Prometheus 監控 |
| `monitoring` | grafana | 3000 | Grafana 儀表板 |

## 📊 CI/CD 流程

### GitHub Actions 工作流程

我們的 CI/CD 管道包含以下工作流程：

#### 1. **ci.yml** - 持續整合
- ✅ **代碼品質檢查**: Lint, Format, Vet
- 🧪 **測試執行**: 單元測試 + 覆蓋率 (≥80%)
- 🛡️ **安全掃描**: Gosec, Trivy, 依賴漏洞檢查
- 🔧 **建置驗證**: 跨平台建置測試
- 🐳 **Docker 建置**: 映像建置和安全掃描
- 🚀 **預覽部署**: PR 專用預覽環境

#### 2. **claude-review.yml** - AI 程式碼審查
- 🤖 **自動審查**: Claude 智能程式碼分析
- 🔍 **安全檢查**: SQL injection, XSS, 路徑遍歷
- ⚡ **效能分析**: N+1 queries, 記憶體洩漏
- 🏗️ **架構檢查**: SOLID 原則, 最佳實踐
- 🔧 **自動修復**: 格式化和小錯誤修正 (可選)

#### 3. **security.yml** - 安全掃描
- 🔒 **Go 安全掃描**: Gosec 靜態分析
- 📦 **依賴掃描**: govulncheck 漏洞檢查
- 🔍 **多維度掃描**: Trivy 檔案系統和配置掃描
- 🐳 **映像掃描**: Docker 映像安全分析
- 📊 **彙總報告**: 完整安全分析報告

#### 4. **release.yml** - 發布部署
- 📦 **建立 Release**: GitHub Release 自動化
- 🏗️ **跨平台建置**: Linux, macOS, Windows (AMD64/ARM64)
- 🐳 **映像發布**: 多架構 Docker 映像
- 🚀 **分階段部署**: Staging → Production
- 📢 **通知整合**: Slack/Discord/Teams 通知

### 品質門檻

| 檢查項目 | 標準 | 動作 |
|----------|------|------|
| 單元測試覆蓋率 | ≥ 80% | 🚫 阻止合併 |
| 重大安全問題 | = 0 | 🚫 阻止合併 |
| 高風險安全問題 | ≤ 2 | ⚠️ 需要審查 |
| 建置成功率 | 100% | 🚫 阻止合併 |
| Lint 問題 | = 0 | 🚫 阻止合併 |

## 🔐 安全最佳實踐

### 容器安全
- ✅ 非 root 使用者執行
- ✅ 最小權限原則
- ✅ 唯讀檔案系統
- ✅ Seccomp 安全配置檔
- ✅ 資源限制和超時設定

### CI/CD 安全
- ✅ 機密資訊使用 GitHub Secrets
- ✅ 最小權限 GitHub Token
- ✅ 依賴漏洞自動掃描
- ✅ 映像簽名和驗證
- ✅ SARIF 安全報告整合

## 🚀 部署策略

### 開發流程
```
feature branch → PR → Claude Review → CI Checks → Merge → Deploy
```

### 發布流程
```
Tag → Build → Security Scan → Staging → Manual Approval → Production
```

### 環境管理
- **Development**: 自動部署，完整測試環境
- **Staging**: PR 預覽，接近生產配置
- **Production**: 手動批准，藍綠部署

## 🛠️ 故障排除

### 常見問題

**Q: Docker 建置失敗？**
```bash
# 清理 Docker 快取
docker system prune -f
.docker/docker-dev.sh clean

# 重新建置
.docker/docker-dev.sh build-dev
```

**Q: 開發環境啟動失敗？**
```bash
# 檢查容器狀態
.docker/docker-dev.sh status

# 查看詳細日誌
.docker/docker-dev.sh logs dev
```

**Q: 測試在容器中失敗？**
```bash
# 進入容器手動除錯
.docker/docker-dev.sh shell

# 在容器內執行測試
make test
```

### 效能調整

**記憶體限制**:
```yaml
# docker-compose.yml
deploy:
  resources:
    limits:
      memory: 2G
```

**並行處理**:
```bash
# 環境變數
MAX_ANALYSIS_WORKERS=4
ANALYSIS_TIMEOUT=300s
```

## 📚 相關文件

- [分支保護設定](.github/branch-protection.md)
- [PR 模板](.github/pull_request_template.md)
- [專案 README](../README.md)
- [架構文件](../docs/architecture/)

## 🤝 貢獻指南

1. **Fork 專案** → 建立 feature branch
2. **本地開發** → 使用 `.docker/docker-dev.sh start-dev`
3. **執行測試** → `.docker/docker-dev.sh test`
4. **建立 PR** → 自動觸發 Claude Review 和 CI
5. **合併後** → 自動部署到測試環境

---

**🎯 目標**: 透過自動化 CI/CD 流程，確保程式碼品質、安全性和可靠的部署過程。

**🤖 AI 增強**: Claude Code 自動審查提供智能程式碼分析和改進建議。