# Telegram API 代理配置指南

## 背景说明

在中国大陆服务器上部署 RSS2TG 机器人时，由于网络限制，无法直接访问 Telegram 官方 API (`api.telegram.org`)。为解决这个问题，本项目支持使用代理服务进行通信。

## 代理服务器选项

### 1. 默认推荐: 查询参数格式

```
TELEGRAM_API_URL=https://fyapi.deno.dev/telegram?url=
```

这种格式最兼容，能正确处理 Telegram API 的 JSON 响应。

### 2. 直接拼接格式

```
TELEGRAM_API_URL=https://fyapi.deno.dev/telegram
```

这种格式简单，但可能会导致 JSON 解析错误，不推荐作为首选。

### 3. 替代参数格式

```
TELEGRAM_API_URL=https://fyapi.deno.dev/telegram?api=
```

某些代理服务使用其他参数名称，如果默认的 `url=` 不工作，可以尝试此格式。

### 4. 自建代理

```
TELEGRAM_API_URL=https://your-own-proxy.com/telegram
```

最可靠的方式是自建代理服务，特别是如果您有海外服务器资源。

## 常见错误及解决方案

### 1. JSON 解析错误

```
错误: 使用自定义 API URL 创建 Bot API 失败: json: cannot unmarshal number into Go value of type tgbotapi.APIResponse
```

**可能原因**:
- 代理服务器返回了不符合 Telegram API 格式的 JSON
- 使用了不正确的代理 URL 格式
  
**解决方案**:
- 使用查询参数格式 `?url=` (推荐)
- 确认代理服务支持 Telegram API 代理

### 2. "connection reset by peer" 错误

```
错误: Post "https://api.telegram.org/bot.../getMe": read tcp ...: connection reset by peer
```

**可能原因**:
- 代理服务器拒绝连接
- 代理URL格式不正确
- 代理服务器暂时不可用

**解决方案**:
- 尝试更换代理服务器
- 调整代理URL格式（查询参数/直接拼接）
- 等待一段时间后重试

### 3. "i/o timeout" 错误

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

## 如何测试代理是否可用

使用以下命令测试代理连接:

```bash
# 测试直接访问
curl -v https://fyapi.deno.dev/telegram

# 测试查询参数格式
curl -v "https://fyapi.deno.dev/telegram?url=https://api.telegram.org/bot123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11/getMe"
```

如果返回类似 Telegram API 的 JSON 响应，表示代理正确配置。

## 代理服务的 HTTPS 和 HTTP

推荐使用 HTTPS 协议的代理服务，这样可以保证通信安全。如果代理服务只支持 HTTP，可以这样配置：

```
TELEGRAM_API_URL=http://your-proxy.com/telegram?url=
```

## 自建代理服务

如果公共代理不稳定，建议自建代理服务。可以在海外服务器上使用 Nginx 或 Cloudflare Workers 等工具创建简单的反向代理。

### Nginx 配置示例 (支持查询参数)

```nginx
server {
    listen 80;
    server_name your-proxy-domain.com;
    
    # 支持直接拼接
    location /telegram/ {
        proxy_pass https://api.telegram.org/;
        proxy_set_header Host api.telegram.org;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # 支持查询参数格式
    location /telegram {
        if ($args ~ "url=(.*)") {
            set $target_url $1;
            proxy_pass $target_url;
            break;
        }
        proxy_pass https://api.telegram.org$request_uri;
        proxy_set_header Host api.telegram.org;
    }
}
```

配置完成后，使用 `https://your-proxy-domain.com/telegram?url=` 作为代理URL。 