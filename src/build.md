# クロスビルド


- Raspberry pi

```
GOOS=linux GOARCH=arm go build -o trjx_transfer_raspberry
```

- centos

```
GOOS=linux GOARCH=amd64 go build -o trjx_transferlinux
```

- intel Macintosh

```
GOOS=darwin GOARCH=amd64 go build -o trjx_transfer_intel_mac
```

- Apple Silicon Macintosh

```
GOOS=darwin GOARCH=arm64 go build -o trjx_transfer_m_mac
```


# 設定

実行ファイルと同じ階層の`setting`ディレクトリ内の`config.json`により動作を切り替えることができる。
