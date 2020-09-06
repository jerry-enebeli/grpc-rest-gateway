/**
Service Package takes care of managing the gRPC services gotten from the uploaded proto file.
*/

package service

import (
	"encoding/json"
	"fmt"
	"github.com/jerry-enebeli/proto-parser/ast"
	"github.com/jerry-enebeli/proto-parser/parser"
	"github.com/mitchellh/mapstructure"
	bolt "go.etcd.io/bbolt"
	"log"
	"os"
	"strings"
	"time"
)

const BUCKETNAME = "Services"

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s Service) db() *bolt.DB {
	err := os.Mkdir("/usr/local/bin/gateway", 0777)
	fmt.Println(err)
	db, err := bolt.Open("/usr/local/bin/gateway/service.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	_ = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(BUCKETNAME))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return db
}

func getProtoDetails(file string) *ast.Ast {
	tokens := parser.NewParser(file).Tokens

	protoDetails := ast.NewAst(tokens)

	protoDetails.GenerateAST()

	return protoDetails
}
func (s Service) CreateService(source string) {

	protoDetails := getProtoDetails(source)

	serviceKey := strings.ToLower(protoDetails.Package + "." + protoDetails.Service.Name)
	data := make(map[string]interface{})
	data["created_at"] = time.Now().Format("2006-01-02 3:4:5 pm")
	data["service_details"] = protoDetails.Service

	service, _ := json.Marshal(data)

	db := s.db()
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKETNAME))
		err := b.Put([]byte(serviceKey), service)
		return err
	})

	if err != nil {
		return
	}

	fmt.Println("service created  ✓")

}

func (s Service) GetAllServices() {
	db := s.db()

	err := db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(BUCKETNAME))

		c := b.Cursor()

		fmt.Println("NAME\t\t\t\t\tCREATED\t\t\t\tMETHODS")
		for k, v := c.First(); k != nil; k, v = c.Next() {
			data := make(map[string]interface{})

			_ = json.Unmarshal(v, &data)

			createdAt := data["created_at"].(string)
			serviceDetails := data["service_details"].(map[string]interface{})

			var service ast.Service
			_ = mapstructure.Decode(serviceDetails, &service)

			fmt.Printf("%s\t\t%v\t\t\t%v\n", k, createdAt, len(service.Methods))
		}
		return nil
	})

	if err != nil {
		return
	}

}

func (s Service) GetServiceMethods(name string) {
	db := s.db()
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BUCKETNAME))
		v := b.Get([]byte(name))

		if string(v) == "" {
			fmt.Println("service not found")
		} else {
			data := make(map[string]interface{})

			_ = json.Unmarshal(v, &data)

			serviceDetails := data["service_details"].(map[string]interface{})

			var service ast.Service
			_ = mapstructure.Decode(serviceDetails, &service)

			fmt.Println("NAME\t\t\tINPUT\t\t\t\tOUTPUT")

			for _, method := range service.Methods {
				fmt.Printf("%s\t\t\t%v\t\t\t\t%v\n", method.Name, method.InputTypeName, method.OutPutTypeName)
			}
		}

		return nil
	})

	if err != nil {
		return
	}

}
