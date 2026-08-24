# BUG_REPRO

基线：`green_base_bug_003`。

服务边界将底层错误转换为仅带文本的新错误，调用方无法再使用 `errors.Is` 保留并识别原始错误身份。

复现命令：

```bash
go test ./internal/service -run '^TestBug003ErrorIdentitySurvivesServiceBoundary$'
```

基线预期现象：测试失败，服务错误丢失其底层错误身份。
