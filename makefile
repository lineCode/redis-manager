.PHONY : all
all : frontend backend

.PHONY : frontend
frontend :
	cd frontend && cnpm install && npm run build

.PHONY : backend
backend :
	go build -o build/redis-manager bin/redis-manager.go
	cp bin/conf.example.yaml build/conf.yaml

.PHONY : clean
clean :
	rm -rf frontend/dist/*
	rm -rf build/*