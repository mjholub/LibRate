alias b := build
alias a := all
alias deps := update_deps
alias fe := build_frontend
alias t := test

target_dir := "$HOME/.local/share/LibRate"

all: tidy copy_libs build build_frontend

copy_libs:
	mkdir -p {{ target_dir }}/lib
	cp -rf lib/* {{ target_dir }}/lib

first_run: all
	sh copy_config.sh
	./LibRate -init -exit
	./LibRate migrate -auto-migrate

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
  go build -o LibRate

test:
	go test -v ./...

clean:
	rm -rvf {{ target_dir }}
# vim: set ft=make :
