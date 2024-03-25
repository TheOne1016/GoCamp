.PHONY: docker
docker:
	# 把上次编译的东西删掉
	@rm webook || true
	@docker rmi -f timezone/webook:v0.0.2
	# 运行一下 go mod tidy, 防止go.sum文件不对，编译失败 hahaha
	@go mod tidy
	# 编译成可执行文件，名字叫做 webook
	# 这里可以随便改这个标签，记得对应的k8s部署里面也要改
	@go build -tags=k8s -o webook .
	@docker build -t timezone/webook:v0.0.2 .

