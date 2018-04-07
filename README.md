# takuzu

Golang takuzu library

[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/McKael/takuzu)
[![license](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/McKael/takuzu/master/LICENSE)
[![Build Status](https://travis-ci.org/McKael/takuzu.svg?branch=master)](https://travis-ci.org/McKael/takuzu)

This repository contains a [Go](https://golang.org/) library package that
provides functions to solve, build or validate takuzu puzzles.

- Mercurial repository: https://hg.lilotux.net/golang/mikael/takuzu/
- Github mirror: https://github.com/McKael/takuzu/

Please read the [godoc documentation](https://godoc.org/github.com/McKael/takuzu) for details.

# CLI demo

This project also contains a command line utility, named `gotak`, to solve or
generate puzzles.
I haven't written the tool documentation yet.

To build the gotak CLI utility, you can use the following command (with Go 1.8+):

```
go get github.com/McKael/takuzu/gotak
```

(If you use the Mercurial repository, you have to update the import path manually.)

Pre-built binaries are available on the
[Github Release page](https://github.com/McKael/takuzu/releases).

The utility was written for personal use and is not very user-friendly, but it
should be reasonably efficient (I've been able to generate boards up to 50x50).
Here are a few examples to get started:

Build new 6x6 puzzle:
```
% gotak --new 6

. . . . . .
0 . . . . 0
. . 1 . . .
. . . . 1 .
1 . 0 0 . .
1 . . . . .
```

Solve the board:
```
% gotak --board ......0....0..1.......1.1.00..1.....

. . . . . .
0 . . . . 0
. . 1 . . .
. . . . 1 .
1 . 0 0 . .
1 . . . . .


0 1 1 0 0 1
0 1 0 1 1 0
1 0 1 1 0 0
0 0 1 0 1 1
1 1 0 0 1 0
1 0 0 1 0 1
```

(You can get the board string with the `--out` flag when generating new puzzles.)

Create a PDF with a takuzu puzzle:
```
% gotak --new 10 --to-pdf /tmp/takuzu.pdf
```

# Online puzzle demo

This library is used by GotakJS, an [online takuzu puzzle game](https://lilotux.net/~mikael/takuzu/),
written in Go using [GopherJS](https://github.com/gopherjs/gopherjs).
(On mobile, works best with Chrome.  On a computer, I've tested it with both Firefox and Chrome.)
