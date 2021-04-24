package main

import (
	"flag"
	"fmt"
	"github.com/hashicorp/consul/api"
	_ "github.com/joho/godotenv/autoload"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const AppVersion = "v0.0.1"

// If you need to change it, please keep the following format:
// ./YOUR_DIR_A/YOUR_DIR_B/.../
const dataDir = "./data/"

var (
	version bool
	backup bool
	restore bool
)

func init()  {
	flag.BoolVar(&version, "v",false,"View current version")
	flag.BoolVar(&backup, "backup",false,"Back up all data to file. File Path:" + dataDir)
	flag.BoolVar(&restore, "restore",false,"Restore all data from the backup file. File Path:" + dataDir)
	flag.Parse()

	// 如果没有输入flag或者输入多个flag
	if flag.NFlag() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if version {
		fmt.Println(AppVersion)
		os.Exit(0)
	}
}

func main()  {
	// Connect Consul
	kv := consulConnect()

	switch {
	case backup:
		consulBackup(kv)
	case restore:
		consulRestore(kv)
	}
}

func consulConnect() *api.KV {
	conf := api.DefaultConfig()
	conf.Address = os.Getenv("Consul_Address")
	conf.Token = os.Getenv("Consul_Token")

	// https://github.com/hashicorp/consul/tree/master/api
	client, err := api.NewClient(conf)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return client.KV()
}

// Back up all data to file
func consulBackup(kv *api.KV)  {
	// Get all Kv pairs
	pairs, _, err := kv.List("", nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Remove Data Directory And All Data
	err = os.RemoveAll(dataDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Create An Empty Data Directory
	err = os.MkdirAll(dataDir, os.ModeDir)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// Create files
	for _, v := range pairs {
		err := os.MkdirAll(dataDir + filepath.Dir(v.Key), os.ModeDir)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}


		f, err:= os.Create(dataDir + v.Key)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		_, err = f.Write(v.Value)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		_ = f.Close()
	}
}

// Restore all data from the backup file
func consulRestore(kv *api.KV)  {
	// Delete all KV pairs
	_, err := kv.DeleteTree("", nil)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	err = filepath.Walk(dataDir, func(path string, info fs.FileInfo, err error) error {
		fi, err := os.Stat(path)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		// If it is not a directory, create data
		if !fi.IsDir() {
			prefix := strings.TrimPrefix(dataDir, "./") // data/
			newPath := strings.TrimPrefix(filepath.ToSlash(path), prefix)

			b, err := os.ReadFile(path)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}

			_, err = kv.Put(&api.KVPair{Key:newPath, Value:b}, nil)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}