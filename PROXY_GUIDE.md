# Telegram API 代理配置指南

## 背景说明

在中国大陆服务器上部署 RSS2TG 机器人时，由于网络限制，无法直接访问 Telegram 官方 API (`api.telegram.org`)。为解决这个问题，本项目支持使用代理服务进行通信。

## 代理服务器选项

### 请求模式说明

本项目支持三种请求模式，通过环境变量 `TELEGRAM_API_MODE` 设置：

| 模式值 | 说明 | 例子 |
|-------|------|------|
| 0 | 直接拼接模式（默认） | `https://tg.ll.sd/bot123456/getMe` |
| 1 | 查询参数模式 | `https://tg.ll.sd?url=https://api.telegram.org/bot123456/getMe` |
| 2 | 移除bot前缀模式 | `https://tg.ll.sd/123456/getMe` |

### 1. 默认推荐: fyapi.deno.dev (模式0)

```
TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
TELEGRAM_API_MODE=0
```

这是一个公共代理服务，通常可以正常工作，但可能会有以下限制：
- 请求速率限制
- 服务可能不稳定
- 某些地区可能无法访问

### 2. 使用查询参数格式 (模式1)

```
TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
TELEGRAM_API_MODE=1
```

某些代理服务使用查询参数格式，如果默认模式不工作，可以尝试此格式。

### 3. 使用移除bot前缀格式 (模式2)

```
TELEGRAM_API_URL=https://tg.ll.sd
TELEGRAM_API_MODE=2
```

一些代理服务使用的URL格式是移除了原始URL中的`/bot`前缀，如果前两种模式不工作，可以尝试此格式。

### 4. 自建代理

```
TELEGRAM_API_URL=http://your-own-proxy.com/telegram
TELEGRAM_API_MODE=0  # 根据您的代理实现选择适当的模式
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
- 尝试更换不同的请求模式 (TELEGRAM_API_MODE=0/1/2)
- 调整代理URL格式
- 等待一段时间后重试
- 尝试其他代理服务

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
- 尝试不同的请求模式
- 联系代理服务提供者
- 尝试自建代理服务

## 各类代理服务器配置示例

### 1. tg.ll.sd

```
TELEGRAM_API_URL=https://tg.ll.sd
TELEGRAM_API_MODE=2
```

### 2. fyapi.deno.dev

```
TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
TELEGRAM_API_MODE=0
```

### 3. telehttps.herokuapp.com

```
TELEGRAM_API_URL=https://telehttps.herokuapp.com
TELEGRAM_API_MODE=1
```

## 如何测试代理是否可用

使用以下命令测试代理连接:

```bash
# 测试直接访问
curl -v https://your-proxy-url

# 测试模式0格式
curl -v https://your-proxy-url/botTESTTOKEN/getMe

# 测试模式1格式 
curl -v "https://your-proxy-url?url=https://api.telegram.org/botTESTTOKEN/getMe"

# 测试模式2格式
curl -v https://your-proxy-url/TESTTOKEN/getMe
```

如果返回非空内容且状态码为 200，表示代理基本可用。

## 自建代理服务

如果公共代理不稳定，建议自建代理服务。可以在海外服务器上使用 Nginx 或 Cloudflare Workers 等工具创建简单的反向代理。

### Nginx 配置示例 (适用于模式0)

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

### Cloudflare Worker 示例 (适用于模式1)

```javascript
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
  const url = new URL(request.url)
  const apiUrl = url.searchParams.get('url')
  
  if (!apiUrl) {
    return new Response('Missing url parameter', { status: 400 })
  }
  
  // 转发请求到 Telegram API
  const response = await fetch(apiUrl, {
    method: request.method,
    headers: request.headers,
    body: request.body
  })
  
  return response
}
```

配置完成后，根据您的实现选择适当的模式和URL格式。 