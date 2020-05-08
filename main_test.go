package main

import (
	"os"
	"testing"
	"time"
)

func prepare() {
	os.Readlink("main.log")
	os.Readlink("panic.log")
	os.Readlink("services.json")
}

func TestMainWithoutPermissions(t *testing.T) {
	os.Chdir("/")
	defer func() {
		r := recover()
		if r == nil {
			t.Error("No error")
		}
	}()
	main()
	time.Sleep(10 * time.Second)
	server.GraceStop()
}

func TestMainWithoutEnv(t *testing.T) {
	prepare()
	defer func() {
		r := recover()
		if r == nil {
			t.Log("No error")
		}
	}()
	main()
	time.Sleep(10 * time.Second)
	server.GraceStop()
}

func TestMainWithoutEnvWithFiles(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Log("No error")
		}
	}()
	main()
	time.Sleep(10 * time.Second)
	server.GraceStop()
	prepare()
}
