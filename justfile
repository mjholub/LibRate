alias b := build
alias a := all
alias deps := update_deps
alias fe := build_frontend
alias t := test

target_dir := "$HOME/.local/share/LibRate"

all: tidy build build_frontend

first_run: all
	sh copy_config.sh
	./LibRate -init -exit
	./LibRate migrate -auto-migrate

write_tags:
	git fetch --tags
	git describe --tags --abbrev=0 > .env

tidy:
	go mod tidy -v

update_deps:
	go get -u ./...

build_frontend:
	# prefer pnpm and use npm as fallback
	if [ -x "$(command -v pnpm)" ]; then \
		cd fe; \
		pnpm install; \
		pnpm run build; \
	else \
		cd fe; \
		npm install; \
		npm run build; \
	fi

build:
  go build -ldflags "-w" -o LibRate

test:
	go test -v ./...

clean:
	rm -rvf {{ target_dir }}
# vim: set ft=make :
