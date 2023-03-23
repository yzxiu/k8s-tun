package client

import (
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func GetUUID(gen bool) string {
	// If do not generated uuid file, return directly
	if !gen {
		return uuid.New().String()
	}
	fs, err := filepath.Abs("uuid")
	if err != nil {
		log.Fatalf("read uuid config err")
	}
	_, err = os.Stat(fs)
	if err != nil {
		return createUUID(fs)
	}
	bs, err := ioutil.ReadFile(fs)
	if err != nil {
		return createUUID(fs)
	}
	if len(bs) < 5 || len(strings.Split(string(bs), "\n")) > 1 ||
		len(strings.Split(string(bs), " ")) > 1 {
		return createUUID(fs)
	}
	return string(bs)
}

func createUUID(fs string) string {
	u := ""
	_ = os.Remove(fs)
	f, err := os.Create(fs)
	defer func(f *os.File) {
		_ = f.Close()
	}(f)
	if err != nil {
		log.Fatalf("create uuid file error")
	} else {
		u = uuid.New().String()
		_, err = f.Write([]byte(u))
		if err != nil {
			log.Fatalf("write uuid file error")
		}
	}
	return u
}
