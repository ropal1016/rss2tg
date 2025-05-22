package main

import (
    "log"
    "os"
    "time"
    "fmt"
    "net/http"
    "strings"
    "io/ioutil"

    "rss2tg/internal/bot"
    "rss2tg/internal/config"
    "rss2tg/internal/rss"
    "rss2tg/internal/storage"
    "rss2tg/internal/stats"
)

type App struct {
    bot        *bot.Bot
    rssManager *rss.Manager
    config     *config.Config
    db         *storage.Storage
}

func NewApp(cfg *config.Config, db *storage.Storage, stats *stats.Stats) (*App, error) {
    bot, err := bot.NewBot(cfg.Telegram.BotToken, cfg.Telegram.Users, cfg.Telegram.Channels, db, cfg, "/app/config/config.yaml", stats)
    if err != nil {
        return nil, err
    }

    rssConfigs := make([]rss.Config, len(cfg.RSS))
    for i, rssCfg := range cfg.RSS {
        rssConfigs[i] = rss.Config{
            URLs:           rssCfg.URLs,
            Interval:       rssCfg.Interval,
            Keywords:       rssCfg.Keywords,
            Group:          rssCfg.Group,
            AllowPartMatch: rssCfg.AllowPartMatch,
        }
    }

    rssManager := rss.NewManager(rssConfigs, db)

    app := &App{
        bot:        bot,
        rssManager: rssManager,
        config:     cfg,
        db:         db,
    }

    bot.SetMessageHandler(app.handleMessage)
    bot.SetUpdateRSSHandler(app.updateRSS)
    rssManager.SetMessageHandler(app.handleMessage)

    return app, nil
}

func (app *App) handleMessage(title, url, group string, pubDate time.Time, matchedKeywords []string) error {
    return app.bot.SendMessage(title, url, group, pubDate, matchedKeywords)
}

func (app *App) updateRSS() {
    rssConfigs := make([]rss.Config, len(app.config.RSS))
    for i, rssCfg := range app.config.RSS {
        rssConfigs[i] = rss.Config{
            URLs:           rssCfg.URLs,
            Interval:       rssCfg.Interval,
            Keywords:       rssCfg.Keywords,
            Group:          rssCfg.Group,
            AllowPartMatch: rssCfg.AllowPartMatch,
        }
    }
    app.rssManager.UpdateFeeds(rssConfigs)
    log.Println("RSS订阅已更新")
}

func (app *App) Start() {
    go app.bot.Start()
    go app.rssManager.Start()
    go app.watchConfig()
}

func (app *App) watchConfig() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            newCfg, err := config.Load("/app/config/config.yaml")
            if err != nil {
                log.Printf("加载配置失败: %v", err)
                continue
            }

            if !app.config.Equal(newCfg) {
                log.Println("检测到配置变更，正在更新...")
                app.config = newCfg
                app.bot.UpdateConfig(newCfg)
                app.updateRSS()
            }
        }
    }
}

// 测试代理连接
func testProxyConnection(proxyURL string) error {
    log.Printf("测试代理连接: %s", proxyURL)
    
    // 创建具有超时的客户端
    client := &http.Client{Timeout: 15 * time.Second}
    
    // 尝试直接访问代理服务器
    resp, err := client.Get(proxyURL)
    if err != nil {
        return fmt.Errorf("访问代理服务器失败: %v", err)
    }
    defer resp.Body.Close()
    
    // 检查状态码
    log.Printf("代理服务器返回状态码: %d", resp.StatusCode)
    if resp.StatusCode < 200 || resp.StatusCode >= 500 {
        return fmt.Errorf("代理服务器返回异常状态码: %d", resp.StatusCode)
    }
    
    // 尝试获取响应内容
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("读取代理响应失败: %v", err)
    }
    
    // 检查响应内容，如果过短或包含错误信息则警告
    bodyStr := string(body)
    if len(bodyStr) < 10 {
        log.Printf("警告: 代理响应内容较短: %s", bodyStr)
    }
    if strings.Contains(strings.ToLower(bodyStr), "error") || 
       strings.Contains(strings.ToLower(bodyStr), "not found") {
        log.Printf("警告: 代理响应可能包含错误信息: %s", bodyStr)
    }
    
    log.Printf("代理服务器测试通过")
    return nil
}

func main() {
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetOutput(os.Stdout)

    log.Println("启动 RSS 到 Telegram 机器人")

    var cfg *config.Config
    var err error

    // 检查代理设置
    if apiURL := os.Getenv("TELEGRAM_API_URL"); apiURL != "" {
        log.Printf("检测到代理 URL 设置: %s", apiURL)
        
        // 测试代理连接
        err := testProxyConnection(apiURL)
        if err != nil {
            log.Printf("警告: 代理服务器测试失败: %v", err)
            log.Printf("将继续尝试启动，但可能无法连接 Telegram API")
        }
    } else {
        log.Println("未设置代理，将直接连接 Telegram API")
    }

    // 首先尝试从环境变量加载配置
    cfg = config.LoadFromEnv()

    // 如果环境变量中没有足够的配置信息，则尝试从配置文件加载
    if cfg.Telegram.BotToken == "" || len(cfg.Telegram.Users) == 0 {
        log.Println("环境变量中配置不完整，尝试从配置文件加载")
        cfg, err = config.Load("/app/config/config.yaml")
        if err != nil {
            log.Fatalf("加载配置失败: %v", err)
        }
    }

    // 打印加载的配置（注意不要打印敏感信息如 bot token）
    log.Printf("加载的配置: Users: %v, Channels: %v", cfg.Telegram.Users, cfg.Telegram.Channels)

    db := storage.NewStorage("/app/data/sent_items.txt")
    stats, err := stats.NewStats("/app/data/stats.yaml")
    if err != nil {
        log.Fatalf("创建统计失败: %v", err)
    }

    app, err := NewApp(cfg, db, stats)
    if err != nil {
        log.Fatalf("创建应用失败: %v", err)
    }

    app.Start()

    log.Println("机器人现在正在运行")

    // 保持应用运行
    select {}
}
