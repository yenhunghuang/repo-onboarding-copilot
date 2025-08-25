# 🛡️ Branch Protection Configuration

本文件說明如何設定 GitHub 分支保護規則，確保程式碼品質和安全性。

## 🎯 分支保護策略

### Main Branch Protection Rules

以下設定需要在 GitHub Repository Settings > Branches 中配置：

#### 1. 基本保護設定
- ✅ **Require a pull request before merging**
  - Required approving reviews: `1`
  - Dismiss stale reviews when new commits are pushed: ✅
  - Require review from code owners: ✅

#### 2. 狀態檢查要求
- ✅ **Require status checks to pass before merging**
  - Require branches to be up to date before merging: ✅
  - **必須通過的檢查項目**:
    ```
    Lint & Format Check
    Unit Tests  
    Security Scan
    Integration Tests
    Claude Code Review
    Build Application
    Docker Build & Security Scan
    ```

#### 3. 額外限制
- ✅ **Require conversation resolution before merging**
- ✅ **Restrict pushes that create files larger than 100MB**
- ✅ **Include administrators** (管理員也需遵守規則)
- ❌ Allow force pushes (禁止強制推送)
- ❌ Allow deletions (禁止刪除分支)

### Develop Branch Protection Rules

開發分支的較寬鬆規則：

#### 1. 基本設定
- ✅ **Require a pull request before merging**
  - Required approving reviews: `1`
  - Dismiss stale reviews: ✅

#### 2. 狀態檢查
- ✅ **Require status checks**:
  ```
  Lint & Format Check
  Unit Tests
  Security Scan
  ```

## 🔧 Quality Gates Implementation

### 自動化品質門檻

我們的 CI/CD Pipeline 實作了多層品質檢查：

#### 第一層：程式碼風格與語法
```yaml
- Lint (golangci-lint)
- Format check (gofmt)
- Vet analysis (go vet)
```

#### 第二層：功能驗證
```yaml
- Unit tests (≥80% coverage)
- Integration tests
- Security tests
```

#### 第三層：安全性檢查
```yaml
- Gosec static analysis
- Dependency vulnerability scan
- Trivy filesystem scan
- Docker image security scan
```

#### 第四層：建置驗證
```yaml
- Cross-platform builds
- Docker image build
- Artifact generation
```

### 品質標準

| 檢查項目 | 門檻標準 | 動作 |
|----------|----------|------|
| **Unit Test Coverage** | ≥ 80% | 🚫 Block merge |
| **Critical Security Issues** | = 0 | 🚫 Block merge |
| **High Security Issues** | ≤ 2 | ⚠️ Review required |
| **Build Success** | 100% | 🚫 Block merge |
| **Lint Issues** | = 0 | 🚫 Block merge |

### Claude Code Review Standards

Claude 自動審查的品質門檻：

#### 🚨 Critical (阻止合併)
- SQL injection vulnerabilities
- Hard-coded secrets
- Path traversal vulnerabilities
- Unsafe file operations
- Memory safety issues

#### ⚠️ Warning (需要審查)
- Performance anti-patterns
- Error handling issues
- Security best practices violations
- Architecture inconsistencies

#### 💡 Suggestion (建議改善)
- Code style improvements
- Performance optimizations
- Better abstractions
- Documentation improvements

## 🚀 Branch Strategy

### Git Flow 分支模型

```
main
├── develop
│   ├── feature/user-authentication
│   ├── feature/docker-optimization
│   └── hotfix/security-patch
└── release/v1.2.0
```

#### Branch Types

1. **main**: 生產環境程式碼
   - 完整的品質檢查
   - 自動部署到生產環境
   - 完整的安全掃描

2. **develop**: 開發整合分支
   - 基本品質檢查
   - 自動部署到測試環境
   - 定期安全掃描

3. **feature/***: 功能開發分支
   - 基本檢查 (lint, test, security)
   - 不自動部署

4. **hotfix/***: 緊急修復分支
   - 快速檢查流程
   - 直接合併到 main 和 develop

5. **release/***: 版本發布分支
   - 完整品質檢查
   - 版本標記和發布準備

## 📋 PR Template

創建 `.github/pull_request_template.md`:

```markdown
## 🎯 變更摘要
<!-- 簡述這個 PR 的主要變更內容 -->

## 📝 變更類型
- [ ] 🐛 Bug 修復
- [ ] ✨ 新功能  
- [ ] 🔧 重構
- [ ] 📚 文件更新
- [ ] 🔒 安全性修復
- [ ] ⚡ 效能改善

## 🧪 測試清單
- [ ] 單元測試已通過
- [ ] 集成測試已通過  
- [ ] 手動測試已完成
- [ ] 安全測試已通過

## 🔍 審查重點
<!-- 請審查者特別注意的地方 -->

## 📷 截圖/影片
<!-- 如有 UI 變更，請提供截圖 -->

## 🔗 相關 Issue
Closes #

## ✅ 檢查清單
- [ ] 程式碼遵循專案規範
- [ ] 已添加必要的測試
- [ ] 文件已更新
- [ ] 無破壞性變更
- [ ] Claude Review 已通過
```

## 🔧 GitHub CLI 快速設定

使用 GitHub CLI 快速設定分支保護：

```bash
# 安裝 GitHub CLI
gh extension install cli/gh-api

# 設定 main 分支保護
gh api repos/$OWNER/$REPO/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"contexts":["Lint & Format Check","Unit Tests","Security Scan","Integration Tests","Claude Code Review","Build Application"]}' \
  --field enforce_admins=true \
  --field required_pull_request_reviews='{"required_approving_review_count":1,"dismiss_stale_reviews":true,"require_code_owner_reviews":true}' \
  --field restrictions=null \
  --field required_conversation_resolution=true
```

## 🚨 緊急程序

### Hotfix 流程

當需要緊急修復生產環境問題時：

1. **建立 hotfix 分支**:
   ```bash
   git checkout -b hotfix/critical-security-fix main
   ```

2. **實施修復並測試**

3. **建立緊急 PR**:
   - 標記為 `hotfix` label
   - 觸發快速 CI 流程
   - 需要至少 1 位資深開發者審查

4. **部署驗證**:
   - 自動部署到 staging
   - 手動驗證修復效果
   - 部署到生產環境

### Override Procedures

在極端情況下，如何繞過分支保護：

1. **臨時停用分支保護** (僅限管理員)
2. **執行緊急修復**
3. **立即重新啟用保護**
4. **補充完整的測試和文件**

## 📊 監控與報告

### 品質指標追蹤

- **PR 合併率**: 目標 > 95%
- **首次審查通過率**: 目標 > 80%  
- **安全問題檢出率**: 追蹤趨勢
- **測試覆蓋率**: 維持 > 80%
- **建置成功率**: 目標 > 99%

### 報告機制

- **每日**: 自動安全掃描報告
- **每週**: 程式碼品質趨勢報告  
- **每月**: 分支保護效果評估
- **每季**: 開發流程改善建議