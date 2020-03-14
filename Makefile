# require: fswatch pstree

PID_FILE = /tmp/task-management-api.pid
GO_FILES = $(wildcard *.go)

start:
	@go run $(GO_FILES) & echo $$! > $(PID_FILE)

stop:
	@-kill `pstree -p \`cat $(PID_FILE)\` | grep -o  '[=\-] [0-9]\+' | sed "s/[- \=]//g" | tr "\n" " " | sed "s/00001 //g"` ||:

restart: stop start
	@printf '%*s\n' "40" '' | tr ' ' - && echo "Updated at" $(shell date) && printf '%*s\n' "40" '' | tr ' ' -

serve: start
	@fswatch -or --event=Updated . | \
	xargs -n1 -I {} make restart

.PHONY: start stop restart serve
