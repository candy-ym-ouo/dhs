# BUG_REPRO

基线：`green_base_bug_010`。

Node 克隆时直接复用 TaskTypes 的底层数组；修改返回的克隆会写回原节点持有的状态。

复现命令：

```bash
go test ./internal/model -run '^TestBug010NodeCloneOwnsTaskTypeStorage$'
```

基线预期现象：测试失败，克隆节点的任务类型修改污染原节点。
