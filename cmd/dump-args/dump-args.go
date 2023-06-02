package main

import (
	"bufio"
	"fmt"
	"gitern/misc"
	"log"
	"os"
	"syscall"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Println(err)
	}

	fmt.Printf("uid (%d) gid (%d)\n", syscall.Getuid(), syscall.Getgid())
	fmt.Printf("euid (%d) egid (%d)\n", syscall.Geteuid(), syscall.Getegid())
	fmt.Println("so.Args: ", os.Args)
	fmt.Println("so.Environ: ", os.Environ())
	usage, err := misc.DiskUsage("/")
	fmt.Println("DiskUsage: ", usage, err)
}
