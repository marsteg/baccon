# Welcome to Bac(kend) con(nection Tester)!

The Baccon Tester is a go webserver that has different handler endpoints to create and delete connections and do simple tests to a backend service (currently only postgresql databases)

## Installation

Download the Git repository and run: 
    go run .



## Available Handlers 
Users are be able to create, list connections via the below described handlers. The handlers with a specific connection-id will send a request to the backend system. 


- GET /postgres/
    - list all currently existing postgress connections including a connection status
- POST /postgres/
    - requires a json body in the request that defines the connection. A UUID will be generated and used as connection-id and the connection will be created and established. Example Request Json:
```
{
    "host": "www.example.com",
    "port": "5432",
    "name": "conn-name",
    "user": "postgres-user",
    "password": "postgres-password",
    "database": "postgres-db"
}
```
- GET /postgres/<connection-id>/write
    - writes to the Database

- GET /postgres/<connection-id>/query
    - query the Database

- DEL /postgres/<connection-id>
    - delete a connection

### Further Planned Handlers:

- GET /kafka/
    - list all currently existing kafka connections including a connection status
- POST /kafka/
    - requires a json body in the request that defines the connection. A UUID will be generated and used as connection-id and the connection will be created and established. We will try to attempt to create a Transcient Queue. Example Request Json:
```
{
    "host": "www.example.com",
    "port": 5432,
    "name": "conn-name",
    "user": "kafka-user",
    "password": "kafka-password",
    "queueName": "queue-name"
}
```
- POST /kafka/<connection-id>
    - requires a json body in the request that describes the message that should be send to the queue of the called connection-id. Example Request Json:
```
{
    "foo": "bar",
    "bar": "foo",
    "foobar": "baccon"
}
```
- GET /kafka/<connection-id>   
    - Will read from the queue, that was created when the connection was created and display it to the user


## Helpful Links/tools

- create a docker postgres: 
    docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=pw -e POSTGRES_USER=admin -e POSTGRES_DB=postgres -d postgres

