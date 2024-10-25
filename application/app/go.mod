module github.com/polo2ro/mailinwhite/app

go 1.23.1

require (
	github.com/gorilla/mux v1.8.1
	github.com/polo2ro/mailinwhite/libs v0.0.0
	github.com/redis/go-redis/v9 v9.7.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace github.com/polo2ro/mailinwhite/libs v0.0.0 => ../libs
