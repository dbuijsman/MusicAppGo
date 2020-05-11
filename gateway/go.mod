module gateway

go 1.14

require (
	general v1.0.0
	github.com/gorilla/mux v1.7.4
	github.com/optiopay/kafka/v2 v2.1.1
)

replace general => ../general
