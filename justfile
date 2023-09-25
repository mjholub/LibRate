alias b := build
alias a := all

target_dir := "$HOME/.local/share/LibRate"

all: tidy copy_libs build build_frontend

copy_libs:
	mkdir -p {{ target_dir }}/lib
	cp -rf lib/* {{ target_dir }}/lib

first_run: all 
	./LibRate -init -exit
	./Librate migrate -auto-migrate

tidy:
	go mod tidy -v

update_deps:
	go get -u ./...

build_frontend:
	cd frontend && pnpm run build

build:
  go build

test:
	go test -v ./...

clean:
	rm -rvf {{ target_dir }}
# vim: set ft=make :
