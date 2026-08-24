# BUG_REPRO

基线：`green_base_bug_007`。

服务层将已经取消的 context 替换为后台 context，心跳操作继续执行时下游无法感知调用已经取消。

复现命令：

```bash
go test -race ./internal/service -count=20 -run '^TestBug007CancelledContextReachesDownstream$'
```

基线预期现象：测试失败，记录到的下游 context 不再带有 `context.Canceled` 状态。
