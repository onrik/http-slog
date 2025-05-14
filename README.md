# httpslog
HTTP transport with slog logger

Usage
```golang
package main

import (
	"net/http"

	"github.com/onrik/httpslog"
)

func main() {
	client := &http.Client{
		Transport: httpslog.New(nil),
	}

	client.Get("http://127.0.0.1")
}
```