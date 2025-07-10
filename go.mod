module tinvest_wa

go 1.24.4

replace ldv/tinvest/users => /mnt/wd/nextCloud/src/go/src/ldv/tinvest/users

replace ldv/tinvest => /mnt/wd/nextCloud/src/go/src/ldv/tinvest

replace ldv/tinvest/operations => /mnt/wd/nextCloud/src/go/src/ldv/tinvest/operations

require (
	github.com/gookit/ini/v2 v2.3.1
	github.com/gorilla/websocket v1.5.3
	ldv/tinvest/operations v0.0.0-00010101000000-000000000000
	ldv/tinvest/users v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/gookit/goutil v0.7.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/term v0.29.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	ldv/tinvest v0.0.0-00010101000000-000000000000 // indirect
)
