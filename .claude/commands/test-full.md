# 完整测试套件

运行后端和前端的所有测试，包含完整的覆盖率报告。

## 后端 Go 测试
```bash
cd backend && go test ./... -v -coverprofile=coverage.out -count=1 && go tool cover -func=coverage.out
```

## 前端 Node.js 测试
```bash
npm test -- --coverage
```
