build:
	go build -race cmd/conductor.go

lint:
	pylint tools/conductorctl.py
	flake8 tools/conductorctl.py
	black tools/conductorctl.py

run:
	sudo ./conductor -c configs/config.json 2>> logs/stderr.log >> logs/stdout.log

all: build run
