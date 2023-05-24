

# ![icon](https://raw.githubusercontent.com/JiangKlijna/auto-deploy/main/html/static/icon.svg) auto-deploy
### auto deploy project or run server shell.


## Installation
### from source code
```bash
git clone github.com/jiangklijna/auto-deploy
cd auto-deploy
go run make/build.go
go build -ldflags "-s -w"
```
### from release
[releases](https://github.com/JiangKlijna/auto-deploy/releases)

## Help
```bash
$ auto-deploy -h
Usage: auto-deploy [-c .yml]
Example: auto-deploy -c config.yml

Options:
  -c string
        set configuration file (default "config.yml")
  -h    this help
  -t    test configuration and exit
  -v    show version and exit
```

## License
Source code in **auto-deploy** is available under the [Apache License 2.0](https://github.com/JiangKlijna/auto-deploy/blob/main/LICENSE).
