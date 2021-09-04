build:
	go build cmd/conductor.go

lint:
	pylint tools/conductorctl.py
	flake8 tools/conductorctl.py
	black tools/conductorctl.py

run:
	sudo ./conductor -c configs/config.json

all: build run
