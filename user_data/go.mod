module user_data

go 1.14

require (
	general v1.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-yaml/yaml v2.1.0+incompatible // indirect
	github.com/gorilla/mux v1.7.4
	github.com/optiopay/kafka/v2 v2.1.1
	github.com/prometheus/client_golang v1.5.1
)

replace general => ../general
