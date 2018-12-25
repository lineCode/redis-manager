# redis-manager
redis manager on browser

## demo
http://fly123.tk:8081/#/dashboard

## enviroment
- linux
- windows // not support tty

## require
- node 
- npm
- cnpm
- golang 1.11 +
- redis-cli // require for tty
- dep
- github.com/jteeuwen/go-bindata

## build at linux
```
git clone https://github.com/cocktail18/redis-manager
make
```

## build at windows
```
git clone https://github.com/cocktail18/redis-manager
cd frontend && cnpm i && npm run build
go build -o build/redis-manager bin/redis-manager.go
```

## setup
```
cp bin/conf.example.yaml build/conf.yaml // modify as you need
cd build && chmod u+x redis-manager && ./redis-manager
visit http://localhost:8081
```

## package download
- https://blog.fly123.tk/dl/redis-manager.linux64.tar.gz
- https://blog.fly123.tk/dl/redis-manager.windows64.tar.gz

