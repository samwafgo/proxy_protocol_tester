# Proxy Protocol v1/v2 测试工具

这是一个用于测试 Proxy Protocol v1 和 v2 的命令行工具，可以向支持 Proxy Protocol 的服务器发送测试请求。

## 🎯 功能特性

- ✅ 支持 Proxy Protocol v1 和 v2
- 🌐 支持 IPv4 和 IPv6
- 🔗 支持 TCP 和 UDP 协议标识
- 🖥️ **交互式界面** - 友好的用户引导
- 📝 命令行界面 - 支持脚本化使用
- 🧪 内置测试服务器用于验证功能
- 🔄 支持重复测试
- 💡 智能故障排除提示

## 📦 编译

```bash
go build -o proxy_tester.exe .
```

## 🚀 使用方法

### 🎮 交互式模式 (推荐)

直接运行程序，享受友好的交互式配置体验：

```bash
# 启动交互式模式
proxy_tester.exe

# 或者使用 go run
go run .
```

交互式模式会逐步引导你配置：
- 📋 Proxy Protocol 版本选择
- 🌐 目标服务器地址和端口
- 📍 源地址信息 (模拟真实客户端)
- 🎯 目标地址信息
- 🔗 协议类型 (TCP/UDP)
- 📨 测试消息配置 (支持自定义域名)
- ⏱️ 超时设置
- ✅ 配置确认和测试执行

#### 交互式示例流程：
```
=== Proxy Protocol 测试工具 ===
欢迎使用交互式配置模式！

请选择 Proxy Protocol 版本 (1 或 2) [默认: 1]: 2
请输入目标服务器地址 [默认: 127.0.0.1]: 192.168.1.10
请输入目标服务器端口 [默认: 8080]: 8080

--- 配置源地址信息 (模拟的真实客户端) ---
请输入源IP地址 [默认: 192.168.1.100]: 10.0.0.1
请输入源端口 [默认: 12345]: 12345

--- 配置目标地址信息 ---
请输入目标IP地址 [默认: 192.168.1.200]: 10.0.0.2
请输入目标端口 [默认: 80]: 80

请选择协议类型 (TCP/UDP) [默认: TCP]: TCP

--- 配置测试消息 ---
请选择测试消息类型:
1. HTTP GET 请求 (默认)
2. 自定义消息
3. 不发送消息
请选择 (1-3) [默认: 1]: 1
请输入 Host 域名 [默认: example.com]: api.myserver.com

请输入连接超时时间(秒) [默认: 10]: 10
```

### 📝 命令行模式

适合脚本化使用和自动化测试：

```bash
# 测试 Proxy Protocol v1
proxy_tester.exe -version=1 -server=192.168.1.10 -port=8080

# 测试 Proxy Protocol v2
proxy_tester.exe -version=2 -server=192.168.1.10 -port=8080
```

#### 完整参数示例

```bash
proxy_tester.exe -version=2 \
  -server=192.168.1.10 \
  -port=8080 \
  -src-ip=10.0.0.1 \
  -src-port=12345 \
  -dst-ip=10.0.0.2 \
  -dst-port=80 \
  -protocol=TCP \
  -message="GET / HTTP/1.1\r\nHost: api.myserver.com\r\n\r\n" \
  -timeout=10
```

#### 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-version` | 1 | Proxy Protocol 版本 (1 或 2) |
| `-server` | 127.0.0.1 | 目标服务器地址 |
| `-port` | 8080 | 目标服务器端口 |
| `-src-ip` | 192.168.1.100 | 源IP地址 (模拟真实客户端) |
| `-src-port` | 12345 | 源端口 |
| `-dst-ip` | 192.168.1.200 | 目标IP地址 |
| `-dst-port` | 80 | 目标端口 |
| `-protocol` | TCP | 传输协议 (TCP/UDP) |
| `-message` | HTTP请求 | 发送的测试消息 |
| `-timeout` | 10 | 连接超时时间(秒) |

### 🧪 内置测试服务器

工具包含一个智能测试服务器，可以接收和解析 Proxy Protocol 头部：

```bash
# 启动测试服务器 (默认端口 8080)
proxy_tester.exe server

# 指定端口启动
proxy_tester.exe server 9090

# 使用 go run 启动
go run . server 8080
```

然后在另一个终端测试：

```bash
# 交互式测试
proxy_tester.exe

# 或命令行测试
proxy_tester.exe -version=1 -server=127.0.0.1 -port=8080
```

## 📋 Proxy Protocol 格式说明

### v1 格式 (文本)
```
PROXY TCP4 192.168.1.100 192.168.1.200 12345 80\r\n
PROXY TCP6 2001:db8::1 2001:db8::2 12345 80\r\n
```

### v2 格式 (二进制)
- 12字节签名: `\x0D\x0A\x0D\x0A\x00\x0D\x0A\x51\x55\x49\x54\x0A`
- 1字节版本和命令: `0x21` (PROXY命令)
- 1字节地址族和传输协议
- 2字节地址信息长度 (大端序)
- 变长地址信息 (IPv4: 12字节, IPv6: 36字节)

## 🎯 测试场景

### 1. 负载均衡器测试
验证 HAProxy、Nginx、F5 等负载均衡器的 Proxy Protocol 支持：
```bash
# 测试 HAProxy
proxy_tester.exe -version=2 -server=haproxy.example.com -port=80

# 测试 Nginx (需要配置 proxy_protocol)
proxy_tester.exe -version=1 -server=nginx.example.com -port=8080
```

### 2. 应用服务器测试
验证应用是否能正确解析客户端真实IP：
```bash
# 测试 Web 应用
proxy_tester.exe -version=1 -server=app.example.com -port=8080 \
  -src-ip=203.0.113.1 -message="GET /api/client-ip HTTP/1.1\r\nHost: app.example.com\r\n\r\n"
```

### 3. 协议兼容性测试
测试服务器对不同版本的支持：
```bash
# 先测试 v1
proxy_tester.exe -version=1 -server=target.example.com -port=8080

# 再测试 v2
proxy_tester.exe -version=2 -server=target.example.com -port=8080
```

### 4. IPv6 支持测试
验证 IPv6 地址的处理：
```bash
proxy_tester.exe -version=2 -server=::1 -port=8080 \
  -src-ip=2001:db8::1 -dst-ip=2001:db8::2
```

## 🔧 故障排除

### 常见问题及解决方案

#### ❌ 连接被拒绝
```
测试失败: dial tcp 192.168.1.10:8080: connectex: No connection could be made...
```
**解决方案:**
- 检查目标服务器是否启动
- 验证端口是否正确
- 检查防火墙设置
- 使用 `telnet` 或 `nc` 验证连接性

#### ❌ 连接超时
```
测试失败: dial tcp 192.168.1.10:8080: i/o timeout
```
**解决方案:**
- 增加超时时间: `-timeout=30`
- 检查网络连通性
- 验证目标地址是否正确

#### ❌ 协议不支持
```
测试失败: read tcp 192.168.1.10:8080: connection reset by peer
```
**解决方案:**
- 确认服务器支持 Proxy Protocol
- 尝试不同版本: `-version=1` 或 `-version=2`
- 检查服务器配置

#### ❌ 地址格式错误
```
❌ 请输入有效的端口号 (1-65535)
```
**解决方案:**
- 检查 IP 地址格式
- 确保端口在有效范围内 (1-65535)
- IPv6 地址使用完整格式

### 🧪 调试技巧

1. **使用内置测试服务器验证工具功能:**
```bash
# 终端1: 启动测试服务器
go run . server 8080

# 终端2: 测试连接
go run . -version=1 -server=127.0.0.1 -port=8080
```

2. **启用详细输出查看协议头部:**
测试服务器会显示接收到的 Proxy Protocol 头部详细信息

3. **使用 Wireshark 抓包分析:**
可以抓取网络包查看实际发送的 Proxy Protocol 数据

## 📊 示例输出

### 成功测试输出
```
=== 测试配置确认 ===
Proxy Protocol 版本: v2
目标服务器: 192.168.1.10:8080
源地址: 10.0.0.1:12345
目标地址: 10.0.0.2:80
协议: TCP
测试消息: "GET / HTTP/1.1\r\nHost: api.myserver.com\r\nConnection: close\r\n\r\n"
超时时间: 10秒
==================================================

🚀 开始测试...

已连接到服务器
发送 Proxy Protocol v2 头部 (28 字节)
头部内容 (hex): 0d0a0d0a000d0a515549540a21110010c0a8016400000000c0a801c8000050
Proxy Protocol 头部发送成功
发送测试消息: "GET / HTTP/1.1\r\nHost: api.myserver.com\r\nConnection: close\r\n\r\n"
等待服务器响应...
收到响应 (85 字节):
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 13

Hello, World!

✅ 测试完成!
```

### 测试服务器输出
```
=== Proxy Protocol 测试服务器 ===
启动服务器在端口 8080...
✅ 服务器已启动，监听端口 8080
等待连接...
----------------------------------------
新连接来自: 127.0.0.1:54321
Proxy Protocol v2 头部 (hex): 0d0a0d0a000d0a515549540a21110010c0a8016400000000c0a801c8000050
版本和命令: 0x21
地址族和传输协议: 0x11
地址信息长度: 16
传输协议: TCP
地址族: IPv4
源地址: 192.168.1.100:12345
目标地址: 192.168.1.200:80
收到数据: GET / HTTP/1.1
收到数据: Host: api.myserver.com
收到数据: Connection: close
连接处理完成
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

---

**快速开始:** 直接运行 `go run .` 体验交互式模式！ 🚀