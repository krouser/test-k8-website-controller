build:
	CGO_ENABLED=0 GOOS=linux go build -o website-controller -a pkg/website-controller.go

image: build
	docker build -t stanley2021/website-controller .

push: image
	docker push stanley2021/website-controller:latest
