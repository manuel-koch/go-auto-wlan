# Development

## Create go file with icon data

[2goarray](github.com/cratonica/2goarray)

```shell
go install github.com/cratonica/2goarray
INPUT=icon.png
OUTPUT=icon.go
echo "//+build linux darwin" > $OUTPUT
$GOPATH/bin/2goarray <VAR_NAME> <PACKAGE> < $INPUT >> $OUTPUT
```