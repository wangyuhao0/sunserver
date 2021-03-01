SET CGO_ENABLED=0
SET GOARCH=amd64
go build -mod=vendor -v -o ./bin/sunserver.exe
xcopy "config" "bin/config" /e /c /y