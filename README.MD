# ClamAV
[![Build Status](https://travis-ci.org/B-V-R/clamav.svg?branch=master)](https://travis-ci.org/B-V-R/clamav)
[![Go Report Card](https://goreportcard.com/badge/github.com/B-V-R/clamav)](https://goreportcard.com/report/github.com/B-V-R/clamav)
[![license](https://img.shields.io/github/license/B-V-R/clamav.svg?style=flat-square)]()

## How to run?

1) Clone `clamav` repository
2) Run `docker-compose up`
3) Build go code with `go build .`
4) Run executable created from above command
5) Download `eicar` sample virus file from `https://secure.eicar.org/eicarcom2.zip`
6) Run the following `CURL` command
   `curl -X POST -i -F "file=@eicarcom2.zip" localhost:8000/scan`

Sample Response

```{"file":"eicar","detected":true,"malware":"Eicar-Test-Signature","av_version":"","metadata":""}```
