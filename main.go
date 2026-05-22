package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    script := flag.String("script", "", "json2swagger or schemaParser")
	sortExample := flag.Bool("sortExample", false, "sort example keys")
	sortSchema := flag.Bool("sortSchema", false, "sort schema keys")
    flag.Parse()

	if *script == "" {
        fmt.Println("Enter script name")
        flag.Usage()
        os.Exit(1)
    }

    switch *script {
    case "json2swagger":
        json2swagger(*sortExample, *sortSchema)
    case "schemaParser":
        schemaParser()
    default:
        fmt.Println("Wrong script name")
        flag.Usage()
        os.Exit(1)
    }
}
