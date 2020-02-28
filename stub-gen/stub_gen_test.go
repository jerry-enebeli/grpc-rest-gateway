package stubgen_test

import (
	"github.com/jerry-enebeli/grpc-rest-gateway/stub-gen"
	"testing"
)

func TestGenerateStub(t *testing.T) {
	exampleProtoFile := "/Users/jerry/Documents/grpc-rest/example/user.proto"
	exampleOutPut := "/Users/jerry/Documents/grpc-rest/proto-clients/"
	
	err := stubgen.GenerateStub(exampleProtoFile, exampleOutPut)

	if err != nil {
		t.Errorf("tet failed with error : %v", err)
	}
}
