# BUG_REPRO

基线：`green_base_bug_004`。

扫描轮次标记使用跨 goroutine 的共享全局时间值，多个扫描 worker 同时运行时存在无锁读写。

复现命令：

```bash
go test -race ./internal/scanner -run '^TestBug004RoundMarkerIsSafeAcrossWorkers$'
```

基线预期现象：race detector 报告 `lastRound` 的数据竞争。
