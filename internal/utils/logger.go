package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func Logger() (*slog.Logger, *os.File) {
	file, err := os.OpenFile("./app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o755)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	outputs := io.MultiWriter(file, os.Stdout)
	logger := slog.New(slog.NewTextHandler(outputs, &slog.HandlerOptions{}))
	return logger, file
}
