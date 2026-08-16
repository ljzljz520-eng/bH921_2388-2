# 婚礼现场图片直播

这是一个纯 Go 后端项目，用 JSON 文件持久化摄影师账号、婚礼直播间、现场照片和导出结果。管理员可以开通摄影师账号并管理所有房间，摄影师可以创建房间、上传照片、删除自己的照片和发起房间导出，亲友可以浏览房间及照片。

## 环境

- Go 1.23.12
- `CGO_ENABLED=0`
- 不需要数据库、前端包或外部网络服务

## 运行

确定性内存演示不会启动网络监听，也不会写文件：

```bash
CGO_ENABLED=0 go run ./cmd/weddinglive -demo
```

启动 HTTP 服务：

```bash
CGO_ENABLED=0 go run ./cmd/weddinglive
```

默认监听 `127.0.0.1:8080`，数据写入 `var/wedding-live.json`。可以用 `-config config.example.json` 指定配置。配置只由进程读取，HTTP 路由不提供静态文件下载。

## API

管理员请求使用 `X-Admin-Token`，摄影师请求使用 `X-Photographer-Token`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| `POST` | `/api/admin/accounts` | 管理员开通摄影师账号 |
| `GET` | `/api/admin/rooms` | 管理员查看全部房间 |
| `DELETE` | `/api/admin/rooms/{roomID}` | 管理员删除房间 |
| `GET` | `/api/rooms` | 亲友浏览房间 |
| `POST` | `/api/rooms` | 摄影师创建房间 |
| `GET` | `/api/rooms/{roomID}/photos` | 亲友浏览现场照片 |
| `POST` | `/api/rooms/{roomID}/photos` | 房主上传 Base64 图片 |
| `DELETE` | `/api/rooms/{roomID}/photos/{photoID}` | 房主删除自己的图片 |
| `POST` | `/api/rooms/{roomID}/exports` | 房主创建图片清单导出 |
| `GET` | `/api/rooms/{roomID}/exports` | 房主查看导出结果 |

固定 fixture 位于 `internal/fixture/demo.json`，默认管理员令牌为 `admin-fixture-token`，fixture 中摄影师令牌为 `photo-token-001`。

## 测试

```bash
CGO_ENABLED=0 go test -count=1 ./...
```

取消导出回归用例会在第二个分片开始时同步取消请求。当前注入版本会继续处理剩余分片并提交完整导出，因此该用例稳定失败；其余业务链路用例应通过。
