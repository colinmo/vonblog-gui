build: build-windows-windows build-linux-windows

build-windows-windows:
	set GOOS=windows&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s -H=windowsgui" -o ../bin/vonpost.exe
	
build-linux-windows:
	set GOOS=linux&&set GOARCH=amd64&&cd src&&go build -ldflags "-w -s" -o ../bin/vonpost

clean:
	cd f:\Dropbox\swap\golang\vonblog\bin && del vonpost
	cd f:\Dropbox\swap\golang\vonblog\bin && del vonpost.exe


spell:
	set GOOS=windows&&set GOARCH=amd64&&set CGO_CFLAGS="-IE:/Laboratory/Utility/aspell-dev-0-50-3-3/include"&&go build -ldflags "-w -s -H=windowsgui" -o ../bin/vonpost.exe
