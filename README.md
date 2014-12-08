Sproxy
======
Ssh proxy writen in golang


## deploy

mkdir src
cp -r sproxy src

export GOPATH=$path_of_src/../
go run run.go

## client
go run client_sproxy.go

## TODO
1.add timeout
2.memory leak debug
