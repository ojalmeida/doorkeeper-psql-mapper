<img src="https://github.com/ojalmeida/doorkeeper-core/blob/main/logo.png?raw=true" alt="drawing" width="400"/>
<br>
<br>

**doorkeeper-psql-mapper** plugin can automatically map a postgresql database and expose it through doorkeeper API endpoint

## Getting Started

### Setting-up database

#### Docker

```bash
docker run -p 5432:5432 --name db -e POSTGRES_PASSWORD=password -d postgres

docker exec -ti db psql -U postgres
```

```sql

CREATE TABLE IF NOT EXISTS Users (User_ID INT PRIMARY KEY , Name VARCHAR, Age INT);

```

### Setting-up plugin

```go
package main

import (
	core "github.com/ojalmeida/doorkeeper-core"
	mapper "github.com/ojalmeida/doorkeeper-psql-mapper"
	"os"
)

var m = mapper.PsqlMapper{}

func main() {

	file, err := os.Open("/tmp/config.json")

	if err != nil {
		panic(err)

	}

	core.SetConfigFile(file)
	core.BindPlugin("^/api/v1/", &m)

	m.Configure(mapper.PsqlMapperConfig{
		DbConnectionString: "user=postgres password=password dbname=postgres sslmode=disable",
		PathPrefix:         "/api/v1/",
	})

	err = m.MapDB()
	if err != nil {
		panic(err)
	}

	core.Start()

}
```

Any doubt about configuration files refer to [doorkeeper-core](https://github.com/ojalmeida/doorkeeper-core)

### Starting application

```bash

go run *.go

```



### Manipulating data

#### Creating

```bash
curl 'http://127.0.0.1:8080/users' -d 'user_id=1' -d 'name=john' -d 'age=24' | jq
```

```json

{
    "status": 201,
    "data": [
        {
            "age": 24,
            "name": "john",
            "user_id": 1
        }
    ]
}

```

#### Updating

```bash
curl 'http://127.0.0.1:8080/users/1' -X PUT -d 'name=Jorge' -d 'age=23' | jq
```

```json

{
    "status": 200,
    "data": [
        {
            "age": 23,
            "name": "jorge",
            "user_id": 1
        }
    ]
}

```

#### Deleting

```bash
curl 'http://127.0.0.1:8080/users/1' -X DELETE | jq

# empty 204 response
```



