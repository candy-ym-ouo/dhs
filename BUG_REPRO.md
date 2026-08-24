# BUG_REPRO

基线：`green_base_bug_009`。

多个扫描 worker 同时执行 Round 时会无锁写入共享的轮次标记，race detector 可以稳定捕获数据竞争。

复现命令：

```bash
go test -race ./internal/scanner -count=20 -run '^TestBug009RoundMarkerIsSafeAcrossWorkers$'
```

基线预期现象：测试失败，并输出 `WARNING: DATA RACE`。
