- 本小程序主要用来实现Geneva的以下四条规则，还可以自定义端口、需要修改的window size的值。
```
"[TCP:flags:SA]-tamper{TCP:window:replace:0}-|"
"[TCP:flags:A]-tamper{TCP:window:replace:0}-|"
"[TCP:flags:PA]-tamper{TCP:window:replace:0}-|"
"[TCP:flags:FA]-tamper{TCP:window:replace:0}-|"
```
- 具体用法：https://www.444.run
增加了两个参数 
-t 超时参数 默认为150
-task 线程参数 默认为128

使用 
apt install golang

export GOOS=linux

export GOARCH=amd64

go build -o test ./

./test -task 128	-p 80,443 
