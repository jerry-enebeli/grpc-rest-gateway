package gateway

import (
	"fmt"
	"github.com/jerry-enebeli/grpc-rest-gateway/pkg/service"
	"github.com/jerry-enebeli/grpc-rest-gateway/tools"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var sourceProtoFile string

var s = service.NewService()

var rootCmd = &cobra.Command{
	Use:   "Gateway",
	Short: "gRPC to REST",
	Long:  `Gateway is a api gateway for gRPC application. gateway maps RESTFUL API to gRPC services.Complete documentation is available at https://github.com/jerry-enebeli/grpc-rest-gateway`,
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := tools.Shell("bash", "gateway --help")

		log.Println(output)
	},
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "manages gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("got this", args)
	},
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "list gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		s.GetAllServices()
	},
}

var serviceListMethodsCmd = &cobra.Command{
	Use:       "list-methods",
	Short:     "list gRPC services methods",
	Example:   "Gateway service list-methods [service name]",
	ValidArgs: []string{"test"},
	Run: func(cmd *cobra.Command, args []string) {
		s.GetServiceMethods(args[0])
	},
}

var serviceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create new gRPC services",
	Run: func(cmd *cobra.Command, args []string) {
		s.CreateService(sourceProtoFile)
	},
}

func addServiceCreateFlags() {
	serviceCreateCmd.Flags().StringVarP(&sourceProtoFile, "source", "s", "", "Source directory to read proto file from")
}

func Execute() {
	addServiceCreateFlags()
	serviceCmd.AddCommand(serviceCreateCmd, serviceListCmd, serviceListMethodsCmd)
	rootCmd.AddCommand(serviceCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
