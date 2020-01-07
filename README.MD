# ClamAV

## How to run?

1) Clone `clamav` repository
2) Run `docker-compose up`
3) Build go code with `go build .`
4) Run executable created from above command
5) Download `eicar` sample virus file from `https://secure.eicar.org/eicarcom2.zip`
6) Run the following `CURL` command
   `curl -X POST -i -F "file=@/Users/bojja/Downloads/go1.13.darwin-amd64.pkg" localhost:8000/scan`

Sample Response

```{"file":"eicar","detected":true,"malware":"Eicar-Test-Signature","av_version":"","metadata":""}```