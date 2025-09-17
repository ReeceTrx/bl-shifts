build-linux:
	mkdir -p out/linux-x64
	GOOS=linux GOARCH=amd64 go build -o out/linux-x64/app ./*.go

build-windows:
	mkdir -p out/windows-x64
	GOOS=windows GOARCH=amd64 go build -o out/windows-x64/app.exe ./*.go

release: clean build-linux build-windows
	tar cvzf out/bl-shifts-${RELEASE_VERSION}-linux-x64.tar.gz --directory out/linux-x64/ app
	zip -j out/bl-shifts-${RELEASE_VERSION}-windows-x64.zip out/windows-x64/*

clean:
	rm -rf out/
