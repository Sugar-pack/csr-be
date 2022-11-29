package logger

import "go.uber.org/zap"

func Get() (*zap.Logger, error) {
	var loggerConfig = zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(zap.DebugLevel)

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
