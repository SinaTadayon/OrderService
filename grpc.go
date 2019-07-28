package main

type CartServer struct{}

//func startGrpcServer() {
//	var port string
//	if App.config.App.Port == "" {
//		port = "9090"
//	} else {
//		port = App.config.App.Port
//	}
//	lis, err := net.Listen("tcp", ":"+port)
//	if err != nil {
//		logger.Err("Failed to listen to TCP on port " + port + err.Error())
//	}
//	logger.Audit("app started at " + port)
//
//	// Start GRPC server and register the server
//	grpcServer := grpc.NewServer()
//	pb.RegisterCartServiceServer(grpcServer, &CartServer{})
//	if err := grpcServer.Serve(lis); err != nil {
//		logger.Err("Failed to listen to gRPC server. " + err.Error())
//	}
//}
