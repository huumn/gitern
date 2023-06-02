package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("       _ _                  ")
	fmt.Println("      (_) |                 ")
	fmt.Println("  __ _ _| |_ ___ _ __ _ __  ")
	fmt.Println(" / _` | | __/ _ \\ '__| '_ \\ ")
	fmt.Println("| (_| | | |_| __/ |  | | | |")
	fmt.Println(" \\__, |_|\\__\\___|_|  |_| |_|")
	fmt.Println("  __/ |                     ")
	fmt.Println(" |___/                      ")
	fmt.Println("                            ")
	fmt.Println("No interactive login allowed")
	os.Exit(128)
}
