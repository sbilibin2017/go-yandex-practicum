package main

// var err error

// // Declare startMetricAgentFunc with exact signature as StartMetricAgent
// var startMetricAgentFunc = func(
// 	ctx context.Context,
// 	serverAddress string,
// 	header string,
// 	key string,
// 	cryptoKeyPath string,
// 	pollInterval int,
// 	reportInterval int,
// 	batchSize int,
// 	rateLimit int,
// ) error {
// 	return workers.StartMetricAgent(
// 		ctx,
// 		serverAddress,
// 		header,
// 		key,
// 		cryptoKeyPath,
// 		pollInterval,
// 		reportInterval,
// 		batchSize,
// 		rateLimit,
// 	)
// }

// func run(ctx context.Context) error {
// 	err = logger.Initialize(logLevel)
// 	if err != nil {
// 		return err
// 	}
// 	defer logger.Sync()

// 	ctx, stop := signal.NotifyContext(
// 		ctx,
// 		syscall.SIGINT,
// 		syscall.SIGTERM,
// 		syscall.SIGQUIT,
// 	)
// 	defer stop()

// 	errCh := make(chan error, 1)

// 	go func() {
// 		errCh <- startMetricAgentFunc(
// 			ctx,
// 			flagServerAddress,
// 			hashKeyHeader,
// 			flagKey,
// 			flagCryptoKey,
// 			flagPollInterval,
// 			flagReportInterval,
// 			batchSize,
// 			flagRateLimit,
// 		)
// 	}()

// 	select {
// 	case <-ctx.Done():
// 		return nil
// 	case err := <-errCh:
// 		return err
// 	}
// }
