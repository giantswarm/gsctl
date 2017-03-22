package commands

import "testing"

func TestExecutable01(t *testing.T) {
	isExec, err := executable("/etc/hosts")
	if err != nil {
		t.Error("Error in executable():", err)
	}
	if isExec != false {
		t.Error("Expected false, got ", isExec)
	}
}

func TestExecutable02(t *testing.T) {
	isExec, err := executable("/bin/echo")
	if err != nil {
		t.Error("Error in executable():", err)
	}
	if isExec != true {
		t.Error("Expected true, got ", isExec)
	}
}

func TestWhich01(t *testing.T) {
	echoPath, err := which("echo")
	if err != nil {
		t.Error("Error in which():", err)
	}
	if echoPath != "/bin/echo" {
		t.Error("Expected '/bin/echo', got ", echoPath)
	}
}

func TestWhich02(t *testing.T) {
	echoPath, err := which("sdfdfhgfdvsbnf")
	if err != nil {
		t.Error("Error in which():", err)
	}
	if echoPath != "" {
		t.Error("Expected '', got ", echoPath)
	}
}
