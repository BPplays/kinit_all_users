package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)


func chown_r (file string, username string, group string) {
	cmd := exec.Command("chown", fmt.Sprintf("%s:%s", username, group), "-R", file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Error running chown for", file, ":", err)
	} else {
		log.Println("chowned", file)
	}
}

func chmod_r (file string, perms string) {
	cmd := exec.Command("chmod", perms, "-R", file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Error running chmod for", file, ":", err)
	} else {
		log.Println("chmoded", file)
	}
}



func main() {
	keytabs_dir := "/var/keytabs"

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	for {

		files, err := os.ReadDir(keytabs_dir)
		if err != nil {
			log.Println("Error reading directory:", err)
			return
		}
	
		for _, file := range files {
			if file.IsDir() {
				dirPath := filepath.Join(keytabs_dir, file.Name())
				keytabFile := filepath.Join(dirPath, file.Name()+".keytab")
	
				if _, err := os.Stat(keytabFile); err == nil {
					filename := file.Name()
	
					chown_r(dirPath, filename, filename)

					chmod_r(dirPath, "700")

					cmd := exec.Command("sudo", "-u", filename, "kinit", fmt.Sprintf("%s@SUZUKO.ORG", filename), "-k", "-t", keytabFile)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
	
					err := cmd.Run()
					if err != nil {
						log.Println("Error running kinit for", filename, ":", err)
					} else {
						log.Println("Ran kinit for", filename)
					}
				}
			}
		}
		
		time.Sleep(3 * time.Hour)

	}
}
