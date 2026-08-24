# BUG_REPRO

基线：`green_base_bug_005`。

Node 克隆时直接复用 TaskTypes 的底层数组；调用方修改克隆结果会污染原节点状态。

复现命令：

```bash
go test ./internal/model -run '^TestBug005NodeCloneOwnsTaskTypeStorage$'
```

基线预期现象：测试失败，克隆节点的修改影响原节点的任务类型。
