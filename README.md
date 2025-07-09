# GitHub Personal Activity Notifier

🧠 **GitHub Personal Activity Notifier** – Công cụ cá nhân hoá giúp theo dõi toàn bộ hoạt động GitHub của bạn, chạy hoàn toàn nền qua GitHub Actions.

## 🎯 Tính năng chính

- 🔔 **Nhắc nhở tức thời**: Mỗi 5 phút check PR/Issue cần xử lý
- 🧠 **Tóm tắt hàng ngày**: Báo cáo hoạt động GitHub lúc 9h sáng
- 📬 **Inbox push**: Theo dõi GitHub Notifications
- 💬 **Gửi thông báo**: Qua Discord webhook
- ✅ **Free 24/7**: Chạy hoàn toàn trên GitHub Actions

## 🚀 Cài đặt

### 1. Fork repository này

### 2. Tạo Discord Webhook
- Vào Discord Server → Settings → Integrations → Webhooks
- Tạo webhook mới và copy URL

### 3. Cấu hình GitHub Secrets
Vào repository Settings → Secrets and variables → Actions, thêm:

| Secret | Mô tả |
|--------|-------|
| `GITHUB_TOKEN` | Personal Access Token với quyền `notifications`, `repo`, `read:org` |
| `DISCORD_WEBHOOK` | URL webhook Discord |

### 4. Tùy chỉnh cấu hình (tùy chọn)
Chỉnh sửa file `.github/workflows/notify.yml`:

```yaml
env:
  CHECK_INTERVAL: '5m'        # Tần suất check (5 phút)
  DAILY_REPORT_TIME: '02:00'  # 9h sáng VN = 2h UTC
  CACHE_FILE: 'cache.json'    # File lưu cache
```

## 🔧 Chạy thử

```bash
# Clone repo
git clone https://github.com/yourusername/gh-notify.git
cd gh-notify

# Cài đặt dependencies
go mod download

# Chạy test
export GITHUB_TOKEN="your_token_here"
export DISCORD_WEBHOOK="your_webhook_url"
export GITHUB_USERNAME="your_username"
go run main.go
```

## 📋 Ví dụ thông báo

### Instant Alert
```
🔔 GitHub Alerts (3 items)

🔍 PRs waiting for your review:
• #123 Fix authentication bug
• #124 Add new feature

⏰ Your PRs need attention:
• #125 Update documentation (2 days old)
```

### Daily Digest
```
🧠 Daily Digest – 2025-07-09

📤 PRs you opened:
• #126 Add message reactions

🔍 Reviews given: 2 reviews completed
🐛 Issues resolved: 1 issue closed
```

## 🛠 Cấu trúc dự án

```
gh-notify/
├── main.go                 # Entry point
├── github/
│   ├── api.go             # GitHub API client
│   ├── checker.go         # Logic kiểm tra alerts
├── notify/
│   ├── discord.go         # Discord webhook
│   ├── format.go          # Format thông báo
├── config/
│   └── config.go          # Cấu hình env
├── cache/
│   └── state.go           # Quản lý cache
└── .github/workflows/
    └── notify.yml         # GitHub Actions
```

## 🔐 Bảo mật

- Không commit secrets vào code
- Sử dụng GitHub Secrets để lưu token
- Cache state được commit tự động để tránh duplicate

## 🐛 Troubleshooting

### Lỗi Authentication
```
Error: failed to get user: 401 Unauthorized
```
→ Kiểm tra `GITHUB_TOKEN` có đúng và có quyền cần thiết

### Lỗi Discord Webhook
```
Error: discord API error: 400 Bad Request
```
→ Kiểm tra `DISCORD_WEBHOOK` URL có đúng format

### Không nhận được thông báo
- Kiểm tra GitHub Actions logs
- Đảm bảo workflow được enable
- Kiểm tra timezone settings

## 📝 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 🤝 Contributing

Pull requests are welcome! Vui lòng mở issue trước khi thực hiện major changes.

---

Made with ❤️ for GitHub productivity
