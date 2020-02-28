package stubgen

import (
	"fmt"
	"github.com/jerry-enebeli/grpc-rest-gateway/utilities"
	"log"
	"path"
)

//GenerateStub generate gRPC stubs based on provided protobuf file
//it saves all generated code in /proto-clients dir
func GenerateStub(protofile, output string) error {

	protofileDir := path.Dir(protofile)
	protoInputFile := path.Base(protofile)

	//create proto outout dir
	protoCompilerCommand := fmt.Sprintf("protoc --proto_path=%s/ %s  --go_out=plugins=grpc:%s", protofileDir, protoInputFile, output)

	log.Println("compiling proto file......", protoCompilerCommand)
	//runs protobuf compiler on a bah shell
	_, cmdError, err := utilities.Shell("bash", protoCompilerCommand)

	if err != nil {
		log.Fatal(err)
	}

	if cmdError != "" {
		log.Fatalf("can not compile proto file: %s", cmdError)
	}

	log.Println("proto file compiled, check the dir proto-clients for generated stubs")

	return nil
}
