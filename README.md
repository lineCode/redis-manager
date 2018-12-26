# redis-manager
redis manager on browser

## demo
http://fly123.tk:8081/#/dashboard

## package download
- https://blog.fly123.tk/dl/redis-manager.linux64.tar.gz
- https://blog.fly123.tk/dl/redis-manager.windows64.tar.gz

## quick start
 ```
 go get -u github.com/cocktail18/redis-manager // or download the bin package
 wget https://blog.fly123.tk/dl/conf.example.yaml
 redis-manager -c conf.example.yaml
 visit http://localhost:8081
 ```

## os
- linux
- windows // not support tty

## development require
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
go build -o build/redis-manager ./
```

## setup
```
cp bin/conf.example.yaml build/conf.yaml // modify as you need
cd build && chmod u+x redis-manager && ./redis-manager -c conf.yaml
visit http://localhost:8081
```


