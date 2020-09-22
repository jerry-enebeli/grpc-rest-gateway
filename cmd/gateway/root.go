package gateway

import (
	"fmt"
	"log"
	"os"

	"github.com/jerry-enebeli/grpc-rest-gateway/pkg/service"
	"github.com/jerry-enebeli/grpc-rest-gateway/tools"
	"github.com/spf13/cobra"
)

var sourceProtoFile string
var sourceJsonFile string
var gRPCBackend string
var gateWayPort string

var rootCmd = &cobra.Command{
	Use:   "gateway",
	Short: "gRPC to REST",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := tools.Shell("bash", "gateway --help")
		log.Println(output)
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "manages gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := tools.Shell("bash", "gateway service --help")
		log.Println(output)
	},
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "list gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		s := service.NewService()
		s.GetAllServices()
	},
}

var serviceListMethodsCmd = &cobra.Command{
	Use:                   "list-methods <service name>",
	Short:                 "list gRPC services methods",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		s := service.NewService()
		s.GetServiceMethods(args[0])
	},
}

var serviceRunCmd = &cobra.Command{
	Use:                   "run [flags...] <service name>",
	Short:                 "Runs a gRPC service",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("service name required. gateway service run <service name>")
			return
		}
		s := service.NewService()
		s.Run(args[0], gRPCBackend, gateWayPort, sourceJsonFile)

	},
}

var serviceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create new gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		s := service.NewService()
		s.CreateService(sourceProtoFile)
	},
}

func addServiceCreateFlags() {
	serviceCreateCmd.Flags().StringVarP(&sourceProtoFile, "source", "s", "", "Source directory to read the proto file from")
}

func addRunServiceFlags() {
	serviceRunCmd.Flags().StringVarP(&sourceJsonFile, "source", "s", "", "Source directory to read the rest to rpc mapper json file from")
	serviceRunCmd.Flags().StringVarP(&gRPCBackend, "backend", "b", "", "Address to the gRPC server")
	serviceRunCmd.Flags().StringVarP(&gateWayPort, "port", "p", "", "Custom port for the gateway")
}

func Execute() {
	addServiceCreateFlags()
	addRunServiceFlags()
	serviceCmd.AddCommand(serviceCreateCmd, serviceListCmd, serviceListMethodsCmd, serviceRunCmd)
	rootCmd.AddCommand(serviceCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
