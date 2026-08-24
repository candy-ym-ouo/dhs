# BUG_REPRO

基线：`green_base_bug_002`。

服务层在收到已取消的 context 后替换为后台 context，导致取消信号无法继续传递到下游存储调用。

复现命令：

```bash
go test ./internal/service -run '^TestBug002CancelledContextReachesDownstream$'
```

基线预期现象：测试失败，已取消的 context 被替换或丢失。
