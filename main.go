package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var white_list_dirs []string

var white_list_dirs_nfs []string


var home_maker_temp_directory string
var uid_directory string


func get_user_from_id(id string) (user.User, error) {
	// get the filesystem type of the folder
	getent_out, err := exec.Command("getent", "passwd", id).Output()
	if err != nil {
		return user.User{}, err
	}

	passwd := strings.Split(string(getent_out), ":")
	gecos := strings.Split(passwd[4], ",")
	// log.Println(len(passwd))
	// log.Println(len(gecos))

	username := passwd[0]
	uid := passwd[2]
	gid := passwd[3]

	real_name := gecos[0]

	home_dir := passwd[5]
	// compare the filesystem type with "nfs"
	return user.User{uid, gid, username, real_name, home_dir}, nil
}



// isNFS returns true if the given folder is on an NFS mounted drive
func isNFS(folder string) (bool, error) {
	// get the filesystem type of the folder
	fsType, err := exec.Command("stat", "-f", "-c", "%T", folder).Output()
	if err != nil {
		return false, err
	}
	// compare the filesystem type with "nfs"
	return strings.TrimSpace(string(fsType)) == "nfs", nil
}


func mk_home_dir(dir string, uid int, gid int) (error) {
	will_make_dir := false
	base_dir := filepath.Join(dir, "..")

	if slices.Contains(white_list_dirs, base_dir) {
		will_make_dir = true
	}

	if slices.Contains(white_list_dirs_nfs, base_dir) {
		nfs_status, err := isNFS(base_dir)
		if err != nil {
			return err
		}

		if nfs_status {
			will_make_dir = true
		} else {
			will_make_dir = false
		}
	}



	if will_make_dir {
		os.Mkdir(dir, 0700)
		os.Chown(dir, uid, gid)
	}
	return nil
}



func make_dirs() {
	err := os.MkdirAll(home_maker_temp_directory, 0755)
	if err != nil {
		log.Fatal(err, " home_maker_temp_directory", " ", home_maker_temp_directory)
	}


	err = os.Chmod(home_maker_temp_directory, 0755)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(uid_directory, 0777)
	if err != nil {
		log.Fatal(err, " uid_directory")
	}


	err = os.Chmod(uid_directory, 0777)
	if err != nil {
		log.Fatal(err)
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
	
					cmd := exec.Command("sudo", "-u", filename, "kinit", fmt.Sprintf("%s@SUZUKO.ORG", filename), "-k", "-t", keytabFile)
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
	
					err := cmd.Run()
					if err != nil {
						log.Println("Error running kinit for", filename, ":", err)
					} else {
						log.Println("Ran kinit for", filename)
						// fmt.Println(time.Now(), ": Ran kinit for", filename)
					}
				}
			}
		}
		
		time.Sleep(3 * time.Hour)

	}
}
