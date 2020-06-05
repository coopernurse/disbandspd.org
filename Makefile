.PHONY: build push

build:
	docker build -t coopernurse/disbandspd .

push:
	docker push coopernurse/disbandspd
