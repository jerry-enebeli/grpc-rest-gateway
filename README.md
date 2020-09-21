


# gRPC To REST
grpc-rest-gateway is a service written in golang that enables application developers convert their Google protocol buffers files to a RESTful gateway that communicates directly to their gRPC servers.

# Background

gRPC is a technology that enables application developer build cross language RPC functions by generate clients and server stubs for various languages. It uses protocol buffers as it's data and service definitions. Protocol buffer is a platform-neutral, extensible mechanism for serializing structured data.

gRPC sounds like a promising tool, but it's adoption rate is not as wide as REST or GraphQl.

grpc-rest-gateway aim to provide an easy to use http gateway that service HTTP+JSON to your gRPC services.

grpc-rest-gateway is a service written in Golang that converts gRPC proto file to rest apis.

grpc-rest-gateway serves as a gateway between a rest client, and a gRPC service.


# Get Started
### Install Binary
Download pre-built binaries from https://github.com/jerry-enebeli/grpc-rest-gateway/releases or build from source.

#### Run Install Command
```bash
$ make install-binary
```

#### Check If It Works
```bash
$ gateway
```

## Creating A Service
Create a service from a proto file. Pass the destination of the proto file as the source flag (-s).
```bash
$ gateway service create -s hello.proto
```

## View All Services
Get a list of all gRPC services.
```bash
$ gateway service list
```

![List All Services](https://res.cloudinary.com/dsxddxoeg/image/upload/v1600656236/Screen_Shot_2020-09-21_at_3.43.40_AM_g4llrn.png)


## View All Methods In A Service
Get a list of all methods in a service.

```bash
$ gateway service list-methods helloworld.greeter
```

![List Service Methods](https://res.cloudinary.com/dsxddxoeg/image/upload/v1600656503/Screen_Shot_2020-09-21_at_3.48.03_AM_kdf7zs.png)

## Run API Gateway for a service 
Create a http API gateway for a gRPC service.
* --backend is the address of the gRPC server
* --port is the custom port for the API gateway

```bash
$ gateway service run helloworld.greeter --backend=127.0.0.1:50051 --port=4300
```

After running the gateway for a service gateway creates a json file which serves a mapper between the gRPC method and custom http routes and method.
It crates a [package].[service].json file e.g helloworld.greeter.json.
```json
{
  "routes": [
    {
      "grpc_path": "/helloworld.Greeter/SayHello",
      "method": "POST",
      "route": "/sayhello"
    }
  ]
}
```

## Run API Gateway for a service with a mapper json file
Create a http API gateway for a grpc service.
* --backend is the address of the grpc server
* --port is the custom port for the API gateway
* -s is the gRPC to REST json file. it defines the mapping between a gRPC service methods, and a custom REST routes.

```bash
$ gateway service run helloworld.greeter --backend=127.0.0.1:50051 --port=4300 -s helloworld.greeter.json
```
The below json is a modification of the generated json file. The json file was generated because a  gRPC to REST mapper file empty in the above command using the -s flag. The file can be named anything, does not have to following the naming convention [package].[service].json but must follow the json structure.

```json
{
  "routes": [
    {
      "grpc_path": "/helloworld.Greeter/SayHello",
      "method": "PUT",
      "route": "/hello"
    }
  ]
}
```

## Register a JSON codec with the gRPC server. In Go, it can be automatically registered simple by adding the following import:

`import _ "github.com/jerry-enebeli/grpc-rest-gateway/codec"`


## Make a http call
Make a http call to the defined http routes in the  gRPC to REST mapper json file. This would send a request to the gRPC server and return the appropriate response from the server.
```bash
curl --location --request PUT 'http://localhost:4300/hello' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "jerry"
}'
```
