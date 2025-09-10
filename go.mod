module tinvest_wa

go 1.24.6

replace ldv/tinvest/users => ../ldv/tinvest/users

replace ldv/tinvest => ../ldv/tinvest

replace ldv/tinvest/operations => ../ldv/tinvest/operations

replace ldv/tinvest/instruments => ../ldv/tinvest/instruments

replace ldv/tinvest/marketdataservice => ../ldv/tinvest/marketdataservice

require (
	github.com/gookit/ini/v2 v2.3.1
	github.com/gorilla/websocket v1.5.3
	ldv/tinvest/marketdataservice v0.0.0-00010101000000-000000000000
	ldv/tinvest/operations v0.0.0-00010101000000-000000000000
	ldv/tinvest/users v0.0.0-00010101000000-000000000000
	modernc.org/sqlite v1.38.2
)

require (
	github.com/HugoSmits86/nativewebp v1.1.0 // indirect
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/tiff v1.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pdfcpu/pdfcpu v0.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/sunshineplan/pdf v1.0.7 // indirect
	golang.org/x/image v0.25.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gookit/goutil v0.7.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/sunshineplan/imgconv v1.1.14
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	ldv/tinvest v0.0.0-00010101000000-000000000000
	ldv/tinvest/instruments v0.0.0-00010101000000-000000000000 // indirect
	modernc.org/libc v1.66.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
