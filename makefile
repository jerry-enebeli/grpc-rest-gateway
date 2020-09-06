PROJECT=gateway
REPO=https://github.com/jerry-enebeli/grpc-rest-gateway
VERSION?="0.0.1"

printProject:
	echo "hello" + ${PROJECT}

install:
	go get ./...


build-binary:
	go build -o ${PROJECT} -pkgdir ${PWD} -a ${PWD}/cmd/main.go

install-binary:
	go install ${PWD}/cmd/gateway.go


