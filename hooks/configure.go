package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("== configure hook runnin")
	cmd := exec.Command("snapctl", "get", "foo")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("== config hook done. res: %q\n", out.String())
	confFile := filepath.Join(os.Getenv("SNAP_COMMON"), "conf")
	touchCmd := exec.Command("touch", confFile)
	err2 := touchCmd.Run()
	if err2 != nil {
		log.Fatal(err2)
	}
}
