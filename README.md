# md5-site-test
## Description
The CLI tool which makes http requests and prints the address of the request along
with the MD5 hash of the response (order is irrelevant). The sites must be fetched in parallel.

## Usage
```
go build -o myhttp main.go
chmod +x myhttp
./myhttp facebook.com yahoo.com https://www.google.com/
```

## CLI Arguments
```
-parallel   max number of requests that can be run in parallel
            (optional, default value: 10)
```