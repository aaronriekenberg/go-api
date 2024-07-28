# go-api

Playing with:
* [h2c server](https://golang.org/x/net/http2/h2c)
* [go 1.22 ServeMux](https://go.dev/blog/routing-enhancements)
* [slog](https://pkg.go.dev/log/slog@latest)
* [xsync.MapOf](https://github.com/puzpuzpuz/xsync)

Handy command for log file viewing:

```
tail -F logs/request.log  | jq '"\(.timestamp) \(.request.host) \(.request.method) \(.request.url) \(.response.code) \(.duration)"'
```
