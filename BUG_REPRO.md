# BUG_REPRO

基线：`green_base_bug_008`。

服务层返回节点查询错误时丢弃了原始错误链；上层无法用 `errors.Is` 判断底层存储错误。

复现命令：

```bash
go test -race ./internal/service -count=20 -run '^TestBug008ErrorIdentitySurvivesServiceBoundary$'
```

基线预期现象：测试失败，服务错误不再匹配底层的 sentinel error。
