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

## build
```
git clone https://github.com/cocktail18/redis-manager
cd redis-manager/bin && dep ensure && go build redis-manager
cd redis-manager/frontend && cnpm install && npm run build
cp -r redis-manager/frontend/dist redis-manager/bin/dist
```

## setup
```
cd redis-manager/bin
mv conf.example.yaml conf.yaml // modify the config file if necessity 
./redis-manager
visit http://localhost:8081
```

## package download
- https://blog.fly123.tk/dl/redis-manager.linux64.tar.gz
- https://blog.fly123.tk/dl/redis-manager.windows64.tar.gz

