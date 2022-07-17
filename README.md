# go-yaml2json

[![Tests](https://ci.codeberg.org/api/badges/6543/go-yaml2json/status.svg)](https://ci.codeberg.org/6543/go-yaml2json)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/codeberg.org/6543/go-yaml2json?status.svg)](https://godoc.org/codeberg.org/6543/go-yaml2json)
[![Go Report Card](https://goreportcard.com/badge/codeberg.org/6543/go-yaml2json)](https://goreportcard.com/report/codeberg.org/6543/go-yaml2json)

<a href="https://codeberg.org/6543/go-yaml2json">
    <img alt="Get it on Codeberg" src="https://codeberg.org/Codeberg/GetItOnCodeberg/media/branch/main/get-it-on-neon-blue.png" height="60">
</a>

golang lib to convert yaml into json

```sh
go get codeberg.org/6543/go-yaml2json
```

```go
yaml2json.Covnert(data []byte) ([]byte, error)
yaml2json.StreamConverter(r io.Reader, w io.Writer) error
```
