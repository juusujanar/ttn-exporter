build:
	go build -o bin/ttn_exporter ttn_exporter.go

run:
	go run ttn_exporter.go

compile:
	GOOS=linux GOARCH=arm     go build -o bin/linux/arm/ttn_exporter ttn_exporter.go
	GOOS=linux GOARCH=arm64   go build -o bin/linux/arm64/ttn_exporter ttn_exporter.go
	GOOS=linux GOARCH=386     go build -o bin/linux/386/ttn_exporter ttn_exporter.go
	GOOS=linux GOARCH=amd64   go build -o bin/linux/amd64/ttn_exporter ttn_exporter.go
	GOOS=darwin GOARCH=arm64  go build -o bin/darwin/arm64/ttn_exporter ttn_exporter.go
	GOOS=darwin GOARCH=amd64  go build -o bin/darwin/amd64/ttn_exporter ttn_exporter.go
	GOOS=freebsd GOARCH=arm64 go build -o bin/freebsd/arm64/ttn_exporter ttn_exporter.go
	GOOS=freebsd GOARCH=amd64 go build -o bin/freebsd/amd64/ttn_exporter ttn_exporter.go
	GOOS=freebsd GOARCH=386   go build -o bin/freebsd/386/ttn_exporter ttn_exporter.go
	GOOS=windows GOARCH=amd64 go build -o bin/windows/amd64/ttn_exporter ttn_exporter.go
