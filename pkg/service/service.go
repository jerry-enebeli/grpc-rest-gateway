/**
service Package takes care of managing the gRPC services gotten from the uploaded proto file.
*/

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jerry-enebeli/grpc-rest-gateway/codec"
	"github.com/jerry-enebeli/grpc-rest-gateway/configs/db"
	"github.com/jerry-enebeli/proto-parser/ast"
	"github.com/jerry-enebeli/proto-parser/parser"
	"github.com/mitchellh/mapstructure"
	"github.com/urfave/negroni"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

type Service interface {
	CreateService(source string) error
	GetAllServices()
	GetService(service string) (packageData, error)
	GetServiceMethods(service string)
	InvokeGrpcMethod(path string, in input) output
	Run(service, backend, port, file string)
}

type RegisterData struct {
	GrpcPath string `json:"grpc_path"`
	Method   string `json:"method"`
	Route    string `json:"route"`
}

type service struct {
	conn     *grpc.ClientConn
	bolt     *db.BoltDB
	register map[string]RegisterData
}

type input map[string]interface{}
type output map[string]interface{}

type packageData map[string]interface{}

func (p packageData) getPackageName(service string) string {
	return strings.Split(service, ".")[0]
}

func (p packageData) getServiceDetails() ast.Service {
	serviceDetails := p["service_details"].(map[string]interface{})
	var service ast.Service
	_ = mapstructure.Decode(serviceDetails, &service)
	return service
}

func NewService() Service {
	register := make(map[string]RegisterData)
	return &service{register: register}
}

func (s service) CreateService(source string) error {
	protoDetails := getProtoDetails(source)
	serviceKey := strings.ToLower(protoDetails.Package + "." + protoDetails.Service.Name)
	createdAt := time.Now().Format("2006-01-02 3:4:5 pm")
	data := input{"created_at": createdAt, "service_details": protoDetails.Service}
	service, _ := json.Marshal(data)

	dbConn := db.NewBoltDB(db.SERVICEBUCKETNAME).Conn

	defer dbConn.Close()

	err := dbConn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.SERVICEBUCKETNAME))
		err := b.Put([]byte(serviceKey), service)
		return err
	})

	if err != nil {
		return err
	}

	fmt.Println("service created  ✓")

	return nil
}

func (s service) GetAllServices() {
	dbConn := db.NewBoltDB(db.SERVICEBUCKETNAME).Conn

	defer dbConn.Close()

	err := dbConn.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(db.SERVICEBUCKETNAME))

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

	return
}

func (s service) GetService(service string) (packageData, error) {
	dbConn := db.NewBoltDB(db.SERVICEBUCKETNAME).Conn

	defer dbConn.Close()
	var serviceData packageData
	err := dbConn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.SERVICEBUCKETNAME))
		v := b.Get([]byte(service))

		if string(v) == "" {
			return errors.New("service not found")
		} else {
			data := packageData{}
			_ = json.Unmarshal(v, &data)
			serviceData = data
		}

		return nil
	})

	if err != nil {
		return packageData{}, err
	}

	return serviceData, nil
}

func (s service) GetServiceMethods(service string) {
	dbConn := db.NewBoltDB(db.SERVICEBUCKETNAME).Conn

	defer dbConn.Close()
	err := dbConn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.SERVICEBUCKETNAME))
		v := b.Get([]byte(service))

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

	return
}

func (s *service) InvokeGrpcMethod(path string, in input) output {
	out := output{}
	err := s.conn.Invoke(context.Background(), path, in, &out)

	if err != nil {
		fmt.Println(err)
	}
	return out
}

func (s *service) Run(service, backend, port, file string) {
	sigs := make(chan os.Signal, 1)

	s.registerService(service, file)
	s.dailGrpcClient(backend)
	go s.startHttpServer(port)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
}

func (s *service) dailGrpcClient(backend string) {
	log.Printf("creating connection to gRPC server at %s ....", backend)
	conn, err := grpc.Dial(backend, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codec.JSON{}.Name())))
	if err != nil {
		log.Println(err)
	}
	s.conn = conn
}

func (s *service) startHttpServer(port string) {
	n := negroni.Classic()
	n.UseHandler(s)

	log.Printf("starting gateway on port %s ....", port)

	err := http.ListenAndServe(":"+port, n)
	if err != nil {
		log.Fatalln(err)
	}

}

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	data, ok := s.register[path]
	if !ok {
		responseJSON(w, 404, nil)
		return
	}
	var in input
	_ = json.NewDecoder(r.Body).Decode(&in)
	out := s.InvokeGrpcMethod(data.GrpcPath, in)
	responseJSON(w, 200, out)
}

func (s *service) registerService(service, file string) {
	packageData, err := s.GetService(service)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if file == "" {
		s.createDefaultRegister(packageData, service)
		return
	}
	s.loadRegisterFromFile(file)
}

func (s *service) loadRegisterFromFile(file string) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
	}

	var fileData map[string][]RegisterData

	_ = json.Unmarshal(data, &fileData)

	registerData := fileData["routes"]

	s.resetRegister()
	for _, rd := range registerData {
		s.register[rd.Route] = rd
	}
}

func (s *service) createDefaultRegister(pd packageData, service string) {
	var r []RegisterData
	packageName := pd.getPackageName(service)
	serviceDetails := pd.getServiceDetails()
	s.resetRegister()
	for _, method := range serviceDetails.Methods {
		rpcPath := fmt.Sprintf("/%s.%s/%s", packageName, serviceDetails.Name, method.Name)
		httpRoute := fmt.Sprintf("/%s", strings.ToLower(method.Name))
		rd := RegisterData{
			GrpcPath: rpcPath,
			Method:   "POST",
			Route:    httpRoute,
		}
		r = append(r, rd)
		s.register[rd.Route] = rd
	}

	fileData := map[string][]RegisterData{"routes": r}

	rJson, err := json.MarshalIndent(&fileData, "", "  ")
	if err != nil {
		fmt.Println(err)
	}

	_ = ioutil.WriteFile(service+".json", rJson, 0644)
}

func (s *service) resetRegister() {
	s.register = map[string]RegisterData{}
}

func getProtoDetails(file string) *ast.Ast {
	tokens := parser.NewParser(file).Tokens

	protoDetails := ast.NewAst(tokens)

	protoDetails.GenerateAST()

	return protoDetails
}

func responseJSON(res http.ResponseWriter, status int, object interface{}) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	err := json.NewEncoder(res).Encode(object)

	if err != nil {
		return
	}
}
