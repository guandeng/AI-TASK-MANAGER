# 单元测试更新

查找并更新未覆盖最新代码变更的单元测试。

## 步骤

### 1. 获取最近修改的代码文件
```bash
git diff --name-only HEAD~5
```

### 2. 运行后端测试，识别失败的测试
```bash
cd backend && go test ./internal/handlers/... -v 2>&1 | grep -E "(FAIL|PASS|RUN)"
```

### 3. 检查测试覆盖率
```bash
cd backend && go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep -E "[0-9]+\.[0-9]+" | awk '$NF < 80 {print}'
```

### 4. 分析需要更新的测试
对比代码变更和测试文件，识别：
- 新增/修改的 Handler 方法没有对应测试
- 新增 API 路由缺少测试覆盖
- 模型字段变更导致测试失效
- 业务逻辑变更需要更新测试用例

### 5. 更新测试文件
为缺失或失效的测试添加/更新测试用例。

### 6. 验证所有测试通过
```bash
cd backend && go test ./... -v -count=1
npm test
```
