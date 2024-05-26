package main

import (
	"bytes"
	"fmt"
	"github.com/mihailtudos/metrickit/pkg/compressor"
	"log"
	"strings"
)

func main() {
	data := []byte(strings.Repeat(`This is a test message`, 20))
	// сжимаем содержимое data
	b, err := compressor.Compress(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d bytes has been compressed to %d bytes\r\n", len(data), len(b))

	// распаковываем сжатые данные
	out, err := compressor.Decompress(b)
	if err != nil {
		log.Fatal(err)
	}
	// сравниваем начальные и полученные данные
	if !bytes.Equal(data, out) {
		log.Fatal(`original data != decompressed data`)
	}
}
