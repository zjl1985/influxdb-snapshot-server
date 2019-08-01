# influxdb-snapshot-server

## windows打包

```shell
.\build.bat
```

## 更新日志

### 1.1.1

- 手动关闭管道，减小内存的占用，本身golang会回收内存

### 1.1.0

- 加入mapstructure来作为对象转换
- 历史查询启用并发
- 实时数据查询如果内存里面没有数据则从数据库里查找


### 1.0.0

- 修改了程序名称`fastdb-snapshot.exe`
- 优化性能，比之前快了30-40%

