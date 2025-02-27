# go-api

Playing with:
* h2c server configured with [go 1.24 http.Protocols](https://pkg.go.dev/net/http@go1.24.0#Protocols)
* [go 1.22 ServeMux](https://go.dev/blog/routing-enhancements)
* [slog](https://pkg.go.dev/log/slog@latest)

Handy command for log file viewing:

```
tail -F logs/request.log  | jq '"\(.timestamp) \(.request.host) \(.request.method) \(.request.url) \(.response.code) \(.duration)"'
```
