package main

import (
	"baize/internal/utils"
	"log/slog"
	"os"
)

//func main() {
//	cpuinfo := storage.GetController()
//	jsonCPU, err := json.MarshalIndent(cpuinfo, " ", "   ")
//	if err != nil {
//		print(err)
//	}
//	fmt.Println(string(jsonCPU))
//}

func main() {
	logger := utils.NewStreamLogger(slog.LevelDebug)
	logger.Debug("debug message")
	logger.Info("info message")
	logge := slog.NewJSONHandler(os.Stdout, nil)
	logger2 := slog.New(logge)
	logger2.Info("info message")
	logger2.Warn("warn message")
}
