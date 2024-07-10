package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)


func chown_r(file string, username string, group string) {
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

func k_dest(user string) {
	cmd := exec.Command("sudo", "-u", user, "kdestroy", "-A")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Error running kdestroy for", user, ":", err)
	} else {
		log.Println("kdestroy:", user)
	}
}

func try_load_file(path string, user string) {
    // file, err := os.Open(filepath.Join(path, "KdKSKif7fBUbAoQvIaMFdOLdluoBDaMjOlazFJ7xYVZpDHzFHzVcbxhc2kA417XaT7RnPq7sj2vAn0woFLwm3Wry7sIoyj60BDiBcFMs9cfsgppXDsXOMK0Ryz5kApkl"))
    // if err != nil {
    //     return
    // }
    // defer file.Close()
	cmd := exec.Command("sudo", "-u", user, "find", "\""+path+"\"", "-maxdepth", "1", "-mindepth", "1", "-print", "-quit")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		log.Println("Error listing first file in home for", user, ":", err)
	}
}

func chmod_r(file string, perms string) {
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
	bad_msg := "kinit: Failed to store credentials: Internal credentials cache error while getting initial credentials"
	keytabs_dir := "/var/keytabs"
	home_dir := "/home"
	var loops int64

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
					loops = 0
					for {
						if loops > 5 {
							break
						}

						filename := file.Name()
		
						chown_r(dirPath, filename, filename)

						chmod_r(dirPath, "700")

						cmd := exec.Command("sudo", "-u", filename, "kinit", fmt.Sprintf("%s@SUZUKO.ORG", filename), "-k", "-t", keytabFile)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
		
						out, err := cmd.CombinedOutput()
						if string(out) == bad_msg {
							k_dest(filename)
							loops += 1
						} else {
							break
						}
						if err != nil {
							log.Println("Error running kinit for", filename, ":", err)
						} else {
							log.Println("Ran kinit for", filename)
						}

						try_load_file(filepath.Join(home_dir, filename), filename)

					}
				}
			}
		}
		
		time.Sleep(3 * time.Hour)

	}
}
