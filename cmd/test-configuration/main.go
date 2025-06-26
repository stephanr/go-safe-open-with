package main

import (
	"fmt"
	"regexp"

	"github.com/stephanr/go-safe-open-with/config"
)

func main() {
	fmt.Println("Test")
	config := config.ReadConfiguration()
	fmt.Println(len(config.Allowed))

	regex := regexp.MustCompile("(?:(?<url>[0-9]))+")

	fmt.Printf("Matches: %v\n", regex.FindAllStringSubmatch("123", -1))
	fmt.Printf("SubexpNames [%v]: %v", regex.NumSubexp(), regex.SubexpNames())
}
