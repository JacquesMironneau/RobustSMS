# RobustSMS
Hybrid solution to send SMS at the minimal cost and the maximum robustness


## Run
```
go run main.go
```

Or
```
go build .
./sms
```


## live reload:

```
go install github.com/cosmtrek/air@latest
~/YOUR\_GO\_PATH/bin/air # or your go path
```

## Personnal use case:

Deploy on RPI by building as
```
ARM=7 GOARCH=arm go build .
```
And then copy the binary to the pi

## TODO:

- [x] SMS cloud service provider comparison
- [x] POC SMS AWS
- [x] RPI SMS Setup
- [x] Setup Go http server
- [ ] RPI Serial configuration 
- [x] Interface from server to rpi
- [x] Deployment