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

## Alternative approach to get ClampShell state via I/O Kit directly

[traversing the I/O registry on Mac OS X (iokit)](https://gist.github.com/JonnyJD/6126680)