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
	"github.com/urfave/negroni"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/grpc"
)

type Service interface {
	CreateService(source string) error
	GetAllServices()
	GetService(service string) (serviceData, error)
	GetServiceMethods(service string)
	InvokeGrpcMethod(path string, in map[string]interface{}) map[string]interface{}
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

type serviceData struct {
	CreatedAt      string      `json:"created_at"`
	ServiceDetails ast.Service `json:"service_details"`
	ServiceName    string      `json:"service_name"`
	PackageName    string      `json:"package_name"`
}

func (s serviceData) getKey() string {
	return strings.ToLower(s.PackageName + "." + s.ServiceName)
}
func (s *serviceData) timeStamp() {
	s.CreatedAt = time.Now().Format("2006-01-02 3:4:5 pm")
}

func NewService() Service {
	register := make(map[string]RegisterData)
	return &service{register: register}
}

func (s service) CreateService(source string) error {
	protoDetails := getProtoDetails(source)

	data := serviceData{
		ServiceDetails: protoDetails.Service,
		ServiceName:    protoDetails.Service.Name,
		PackageName:    protoDetails.Package,
	}
	data.timeStamp()
	serviceKey := data.getKey()

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
			var sd serviceData

			_ = json.Unmarshal(v, &sd)

			fmt.Printf("%s\t\t%v\t\t\t%v\n", k, sd.CreatedAt, len(sd.ServiceDetails.Methods))
		}
		return nil
	})

	if err != nil {
		return
	}

	return
}

func (s service) GetService(service string) (serviceData, error) {
	dbConn := db.NewBoltDB(db.SERVICEBUCKETNAME).Conn

	defer dbConn.Close()
	var sd serviceData
	err := dbConn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.SERVICEBUCKETNAME))
		v := b.Get([]byte(service))

		if string(v) == "" {
			return errors.New("service not found")
		} else {
			data := serviceData{}
			_ = json.Unmarshal(v, &data)
			sd = data
		}

		return nil
	})

	if err != nil {
		return serviceData{}, err
	}

	return sd, nil
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
			var sd serviceData

			_ = json.Unmarshal(v, &sd)

			fmt.Println("NAME\t\t\tINPUT\t\t\t\tOUTPUT")

			for _, method := range sd.ServiceDetails.Methods {
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

func (s *service) InvokeGrpcMethod(path string, in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
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
	var in map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&in)
	out := s.InvokeGrpcMethod(data.GrpcPath, in)
	responseJSON(w, 200, out)
}

func (s *service) registerService(service, file string) {
	serviceData, err := s.GetService(service)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if file == "" {
		s.createDefaultRegister(serviceData, service)
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

func (s *service) createDefaultRegister(sd serviceData, service string) {
	var r []RegisterData

	s.resetRegister()
	for _, method := range sd.ServiceDetails.Methods {
		rpcPath := fmt.Sprintf("/%s.%s/%s", sd.PackageName, sd.ServiceName, method.Name)
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
