# BUG_REPRO

基线：`green_base_bug_001`。

注册请求未初始化 Labels 时，写入来源标签会对 nil map 赋值并导致运行时 panic。

复现命令：

```bash
go test ./internal/model -run '^TestBug001RegisterMetadataHandlesNilLabels$'
```

基线预期现象：测试失败，并报告注册元数据处理 nil Labels 时发生 panic。
