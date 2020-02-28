


# gRPC To REST
grpc-rest is a service written in golang that enables application developers convert their Google protocol buffers files to a RESTful gateway that communicates directly to their gRPC services.

# Background

gRPC is a technology that enables application developer build cross language RPC functions by generate clients and server stubs for various languages. It uses protocol buffers as it's data and service definitions. Protocol buffer is a platform-neutral, extensible mechanism for serializing structured data.

gRPC sounds like a promising tool but it's adoption rate is not as wide as REST or GraphQl.

gRPC to Rest aim to provide an easy to use proxy generator that service HTTP+JSON to your gRPC services.

grpc to rest is a service written in Golang that converts grpc proto file to rest apis.

grpc to rest serves as a gateway between a rest client and a grpc service.

# Key Modules

1. gRPC Stub generator
2. gRPC Service Register
3. RESTful Proxy
4. RESTful swagger documentation generator