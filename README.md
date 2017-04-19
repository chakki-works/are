# bowme

Help you to find the commands that you remember ambiguously.

<img src="https://github.com/chakki-works/bowme/raw/master/docs/bowme.png" width="100">

*photo by [DaPuglet](https://flic.kr/p/Q2rT5L)*

## Usage

```
bowme "your ambiguous words"
```

![usage.png](./docs/usage.png)


The default index is get from [bowme.csv](https://gist.github.com/icoxfog417/55cddaa1b0c35c26cac0bace2f2b6940) on public gist.

This file is stored into user home directory `$HOME/.bowme.csv`. 
If you want to add/modify it, please edit the `$HOME/.bowme.csv`.


## Install

If you use Mac/Windows, you can use binary file in the `binary` folder.

Of course, you can build & install yourself.

1. `git clone https://github.com/chakki-works/bowme.git`
2. `go install`

If you encounter `no install location for directory~` error, please try to set `GOBIN`.

```
export GOBIN=$GOPATH/bin
```

