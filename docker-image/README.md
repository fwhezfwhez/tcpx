docker-image

## Note
docker-image provides service exampled in /examples/sayHello, gateway in /gateway/pack-transfer, validation in /all-language-clients all together.Users can spare go environment running these server.

## Step
```
1. go build main.go -o main
2. docker build -t tcpx:latest .
3. docker run -idt --rm -p 7000:7000 -p 7001:7001 -p 7171:7171 -p 7172:7172 -p 7173:7173 tcpx:latest
```
