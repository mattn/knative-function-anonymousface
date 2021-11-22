# knative-function-anonymousface

![](https://raw.githubusercontent.com/mattn/knative-function-anonymousface/main/screenshot.png)

## Usage

```
$ curl -s -X POST -F image=@input.jpg http://anonymousface.default.127.0.0.1.nip.io:8080 > out.jpg
```

## Installation

```
$ kn func deploy
$ kn func run
```

## License

MIT

`data/facefinder` is provided from pigo

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
