BUILD_DATE := `date +%Y-%m-%d\ %H:%M`
VERSIONFILE := version.go
APP_VERSION := `bash ./generate_version.sh`
APP_NAME := "warmind-network"
PROFILE_BENCHMARK_NAME := "BenchmarkMaxLight"
#PROFILE_BENCHMARK_NAME := "BenchmarkFixupProfileFromProfileResponse"


all: build
easyjson:
	easyjson -omit_empty -all bungie/*.go
genversion:
	rm -f $(VERSIONFILE)
	@echo "package main" > $(VERSIONFILE)
	@echo "const (" >> $(VERSIONFILE)
	@echo "  version = \"$(APP_VERSION)\"" >> $(VERSIONFILE)
	@echo "  buildDate = \"$(BUILD_DATE)\"" >> $(VERSIONFILE)
	@echo ")" >> $(VERSIONFILE)
build: genversion
	go build
install: genversion
	go install
test:
	go test -v ./...
bench:
	# Don't run regular tests as part of benchmarks
	go test -v -bench=. -run=XXX ./...
profile:
	go test -v -run=xxx -bench=$(PROFILE_BENCHMARK_NAME) -memprofile mem.out -cpuprofile cpu.out ./bungie
coverage:
	## Right now the coverprofile option is not allowed when testing multiple packages.
	## this is the best we can do for now until writing a bash script to loop over packages.
	go test -cover ./...
#	go test --coverprofile=coverage.out
#	go tool cover -html=coverage.out
minimal: genversion
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o warmind-network .
deploy: genversion
	GOOS=linux GOARCH=amd64 go build
	scp ./$(APP_NAME) li:
	rm $(APP_NAME)
clean:
	rm $(APP_NAME)
