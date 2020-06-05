export PATH=$PATH:$HOME/go/bin

protoc --proto_path=proto --proto_path=third_party --go_out=plugins=grpc:proto chatserver.proto

build: docker build . -t twinklemehta/chatserver:chatserver
run: docker run -it -p 8080:8080 twinklemehta/chatserver:chatserver