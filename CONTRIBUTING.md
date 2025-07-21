# Contributing to Portfolio Management System

## 📝 Git Commit Guidelines

我们使用[Conventional Commits](https://www.conventionalcommits.org/)规范来保持commit历史的清晰和一致性。

### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型

| Type | 描述 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(auth): add user login system` |
| `fix` | Bug修复 | `fix(api): resolve database connection issue` |
| `docs` | 文档变更 | `docs: update README with new setup instructions` |
| `style` | 代码风格调整 | `style(frontend): format components with prettier` |
| `refactor` | 重构代码 | `refactor(services): optimize database queries` |
| `perf` | 性能优化 | `perf(api): improve response time by 30%` |
| `test` | 测试相关 | `test(portfolio): add unit tests for calculations` |
| `build` | 构建系统变更 | `build: update docker configuration` |
| `ci` | CI配置变更 | `ci: add automated testing workflow` |
| `chore` | 其他变更 | `chore: update dependencies` |
| `revert` | 回滚提交 | `revert: revert commit abc1234` |

### Scope 范围 (可选)

表示变更影响的模块或功能：

- `frontend` - 前端相关
- `api` - API网关
- `db` - 数据库相关
- `auth` - 认证系统
- `portfolio` - 投资组合功能
- `market` - 市场数据
- `analytics` - 分析功能
- `docker` - 容器化配置

### Subject 主题

- 使用祈使语气，现在时态
- 不要大写首字母
- 不要以句号结尾
- 限制在50个字符以内

### Body 正文 (可选)

- 解释变更的原因和内容
- 每行限制72个字符
- 可以包含多段

### Footer 页脚 (可选)

- 包含BREAKING CHANGES
- 关联的issue编号

## 🔧 设置Git Commit规范

### 1. 安装依赖并配置

```bash
# 安装commitlint和husky
pnpm install

# 设置Git hooks和模板
pnpm run setup:git
```

### 2. 使用commit模板

Git已配置使用commit模板，执行`git commit`时会自动显示规范格式。

### 3. 自动验证

每次commit时，husky会自动验证commit message格式：

```bash
git add .
git commit -m "feat(frontend): add portfolio dashboard"
```

如果格式不正确，commit会被拒绝并显示错误信息。

## ✅ 示例 Commits

### 好的提交消息

```bash
feat(auth): implement JWT authentication system

Add user login and registration endpoints with JWT token generation.
Includes middleware for protected routes and token validation.

Closes #123
```

```bash
fix(api): resolve memory leak in market data service

The WebSocket connections were not properly closed, causing memory usage
to grow over time. Added proper cleanup in connection handlers.
```

```bash
docs: update setup instructions for Docker

- Add prerequisites section
- Update environment variables
- Include troubleshooting guide
```

### 不好的提交消息

```bash
❌ update stuff
❌ Fix bug
❌ WIP: working on new feature
❌ feat: add new feature that does a lot of things and this description is way too long for a commit subject line
```

## 🚀 开发工作流

1. **创建功能分支**
   ```bash
   git checkout -b feat/portfolio-dashboard
   ```

2. **进行开发**
   ```bash
   # 开发代码...
   ```

3. **提交变更**
   ```bash
   git add .
   git commit -m "feat(frontend): add portfolio dashboard"
   ```

4. **推送分支**
   ```bash
   git push origin feat/portfolio-dashboard
   ```

5. **创建Pull Request**

## 📋 Pull Request Guidelines

- 使用描述性的PR标题
- 填写PR模板中的所有必要信息
- 确保所有测试通过
- 请求至少一个人的代码审查
- 在合并前解决所有评论

---

**注意**: 这些规范通过Git hooks自动执行，无法绕过。如果需要临时跳过验证，请联系项目维护者。