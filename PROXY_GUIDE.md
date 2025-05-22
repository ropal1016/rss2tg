# Telegram API 代理配置指南

## 背景说明

在中国大陆服务器上部署 RSS2TG 机器人时，由于网络限制，无法直接访问 Telegram 官方 API (`api.telegram.org`)。为解决这个问题，本项目支持使用代理服务进行通信。

## 代理服务器选项

### 1. 默认推荐: fyapi.deno.dev

```
TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
```

这是一个公共代理服务，通常可以正常工作，但可能会有以下限制：
- 请求速率限制
- 服务可能不稳定
- 某些地区可能无法访问

### 2. 查询参数格式

```
TELEGRAM_API_URL=http://fyapi.deno.dev/telegram?api=
```

某些代理服务使用查询参数格式，如果默认方式不工作，可以尝试此格式。

### 3. 自建代理

```
TELEGRAM_API_URL=http://your-own-proxy.com/telegram
```

最可靠的方式是自建代理服务，特别是如果您有海外服务器资源。

## 故障排除

### 1. "connection reset by peer" 错误

```
错误: Post "https://api.telegram.org/bot.../getMe": read tcp ...: connection reset by peer
```

**可能原因**:
- 代理服务器拒绝连接
- 代理URL格式不正确
- 代理服务器暂时不可用

**解决方案**:
- 尝试更换代理服务器
- 调整代理URL格式（直接拼接/查询参数）
- 等待一段时间后重试

### 2. "i/o timeout" 错误

```
错误: Post "https://api.telegram.org/bot.../getMe": dial tcp ...: i/o timeout
```

**可能原因**:
- 代理服务器网络延迟高
- 防火墙限制
- 代理服务暂时不可用

**解决方案**:
- 检查网络连接
- 调整超时设置
- 尝试不同的代理服务

### 3. 状态码 4xx/5xx 错误

```
错误: 代理服务器返回异常状态码: 403/404/500等
```

**可能原因**:
- 代理服务拒绝请求
- 代理服务器内部错误
- 请求格式不符合代理要求

**解决方案**:
- 检查代理URL和格式
- 联系代理服务提供者
- 尝试自建代理服务

## 如何测试代理是否可用

使用以下命令测试代理连接:

```bash
curl -v http://fyapi.deno.dev/telegram
```

如果返回非空内容且状态码为 200，表示代理基本可用。

## 自建代理服务

如果公共代理不稳定，建议自建代理服务。可以在海外服务器上使用 Nginx 或 Cloudflare Workers 等工具创建简单的反向代理。

### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name your-proxy-domain.com;
    
    location /telegram/ {
        proxy_pass https://api.telegram.org/;
        proxy_set_header Host api.telegram.org;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

配置完成后，使用 `http://your-proxy-domain.com/telegram` 作为代理URL。 