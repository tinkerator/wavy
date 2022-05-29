# wavy - a tool to convert ASCII art into logic trace images

## Overview

The `wavy` tool is a lightweight program to convert simple ASCII art
logic trace representations into a `PNG`. It was written because it
has very light dependencies and can be used on any platform Go
supports.

## Getting started

Build from source:
```
$ go build wavy.go
```

The `wavy` program transforms `.wvy` files into `.PNG` images. For
example, in the source repository, `examples/hello.wvy` looks like
this:
```
+0 clk
+1,.25 aclk

zzzzzxxx^^^^______zz a
xx_^_^_^_^_^_^_^_^xx b
__/^^^\_____________ c
xx<-><->xxx<->xxxxxx d NOP,ACK,SYN

```

You can invoke `./wavy` against it as follows:
```
$ ./wavy --input=examples/hello.wvy --output=hello.png
```

The generated `hello.png` looks like this:

![Generated example.png image from example/hello.wvy](hello.png "hello.wvy translated to a PNG")

## License info

The `wavy` program is distributed with the same BSD 3-clause license
as that used by [golang](https://golang.org/LICENSE) itself.

## Reporting bugs and feature requests

The `wavy` has been developed purely out of self-interest and a
curiosity for maintaing some project documentation. Should you find a
bug or want to suggest a feature addition, please use the [bug
tracker](https://github.com/tinkerator/wavy/issues).
