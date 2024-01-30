package main

import (
	"encoding/json"
	"os"

	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo"
	"github.com/SkaceKamen/valetudo-telegram-bot/pkg/valetudo_map_renderer"
)

func main() {
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	data, err := os.ReadFile(inputFile)

	if err != nil {
		panic(err)
	}

	var mapData valetudo.RobotStateMap

	err = json.Unmarshal(data, &mapData)

	if err != nil {
		panic(err)
	}

	rendered := valetudo_map_renderer.RenderMap(&mapData)

	err = os.WriteFile(outputFile, rendered, 0644)

	if err != nil {
		panic(err)
	}
}
