# Ryde developer test

## Setup environment
Install Docker
clone project
`git clone https://github.com/rydeapplicant/ryde`
Run mongo container, comment out volumes section in compose file if you do not want to persist DB data.
`docker compose up -d mongo`

## Build and run
```
go build
source .env
./ryde

import postman collection and test endpoints
```

## Test
`go test -v`
