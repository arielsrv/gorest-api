[![pipeline status](https://gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/badges/main/pipeline.svg)](https://gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/-/commits/main)
[![coverage report](https://gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/badges/main/coverage.svg)](https://gitlab.com/iskaypetcom/digital/oms/api-core/gorest-api/-/commits/main)

> gorest-api

# Local

```shell
task build run
```

# Docker
```shell
task docker:build docker:run
```

# Kubernetes

## Create certificates (first time)
```shell
task k:tls
```

## Add a new host entry to host file (first time)
```shell
127.0.0.1 gorest-api.local.dp.iskaypet.com
```

```shell
task k:run
```
