### storage-model

*storage-model* is a storage contract model based on [mitum](https://github.com/imfact-labs/mitum2).

#### Installation

```sh
$ git clone https://github.com/imfact-labs/storage-model

$ cd storage-model

$ go build -o ./imfact ./main.go
```

#### Run

```sh
$ ./imfact init --design=<config file> <genesis config file>

$ ./imfact run --design=<config file>
```

[standalong.yml](standalone.yml) is a sample of `config file`.
[genesis-design.yml](genesis-design.yml) is a sample of `genesis config file`.