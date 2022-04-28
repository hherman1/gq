[![GoDoc](https://godoc.org/github.com/hherman1/gq?status.svg)](https://pkg.go.dev/mod/github.com/hherman1/gq/gq)

# gq
jq but using go instead.


## Install

```
go install github.com/hherman1/gq@latest
```

Or download the binaries [here](https://github.com/hherman1/gq/releases/latest)

# Usage

Pipe some input into `gq` and pass in a block of go code as the argument. The variable `j` is set to the passed in JSON and printed at the end of the function, so manipulate it (e.g below) in order to change the output. The `gq` library is inlined in the program, and described in detail in the API reference above.

# Examples

Accesses

```
$ echo '{"weird": ["nested", {"complex": "structure"}]}' | gq 'j.G("weird").I(1).G("complex")'
"structure"
```

Filters

```
$ echo '["list", "of", "mostly", "uninterseting", "good/strings", "good/welp"]' | gq 'j.Filter(func(n *Node) bool { return strings.Contains(n.String(), "good/") })'
[
	"good/strings",
	"good/welp"
]
```

Maps

```
$ echo '[1, 2, 3, 4]' | gq 'j.Map(func(n *Node) *Node {n.val = n.Int() + 50 / 2; return n})'
[
	26,
	27,
	28,
	29
]
```

# Why?

`jq` is hard to use. There are alternatives like `fq` and `zq`, but they still make you learn a new programming language. Im tired of learning new programming languages.

`gq` is not optimized for speed, flexibility or beauty. `gq` is optimized for minimal learning/quick usage. `gq` understands that you don't use it constantly, you use it once a month and then forget about it. So when you come back to it, `gq` will be easy to relearn. Just use the builtin library just like you would any other go project and you're done. No unfamiliar syntax or operations, or surprising limits. Thats it.


# How's it work?

Speaking of surprising limits, `gq` runs your code in the [yaegi](https://github.com/traefik/yaegi) go interpreter. This means that it runs quickly for small inputs/programs (the alternative was `go run`, which is... not quite as quick). However, it also means its not the fastest `*q` out there, and further it means that you might run into quirks with `yaegi` interpretation limitations. Seems to work pretty well so far though.

# This tool sucks.

Yea, well, I built it last night, so, yea it kinda sucks. But please file issues! Or submit PRs! Maybe this will turn into something useful? At least maybe it will inspire some conversation.

I think it needs a lot of refinement still, in short.
