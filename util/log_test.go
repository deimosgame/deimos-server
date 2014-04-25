package util

import (
	"os"
	"testing"
)

var (
	TestLogFile = "test.log"
	logTest     *Logger
)

func TestInitLogging(t *testing.T) {
	logTest = InitLogging(TestLogFile)
	if _, err := os.Stat(TestLogFile); os.IsNotExist(err) {
		t.Log(`Failed to create the log file. Check file system permissions
			of the current directory`)
		t.Fail()
	}
}

func TestLogging(t *testing.T) {

	if os.Getenv("WERCKER_STEP_ID") != "" {
		t.Log("Wercker detected - skipping log file testing")
		t.Skipped()
		return
	}

	// Debug logging test
	logTest.Debug("This is a debug log test")
	fi, err := os.Stat(TestLogFile)
	if err != nil {
		t.Fail()
	}
	if fi.Size() != 0 {
		t.Log("The debug function seems to write data to log file")
		t.Fail()
	}

	// Error logging test
	logTest.Error("This is an error log test")
	fi, err = os.Stat(TestLogFile)
	if err != nil {
		t.Fail()
	}
	errSize := fi.Size()
	if errSize == 0 {
		t.Log("The error function does not write data to log file")
		t.Fail()
	}

	// Info logging test
	logTest.Info("This is an info log test")
	fi, err = os.Stat(TestLogFile)
	if err != nil {
		t.Fail()
	}
	if fi.Size() != errSize {
		t.Log("The info function unexpectedly wrote data to output file")
		t.Fail()
	}

	// Info logging test to file
	logTest.ToFile = true
	logTest.Info("This is an info log test #2")
	fi, err = os.Stat(TestLogFile)
	if err != nil {
		t.Fail()
	}
	if fi.Size() == errSize {
		t.Log("The info function did not write to the log file")
		t.Fail()
	}

	logTest.Close()
	err = os.Remove(TestLogFile)
	if err != nil {
		t.Log("Error removing the test logging file")
		t.Fail()
	}
}

func TestColorSeq(t *testing.T) {
	if colorSeq(colorBlack) != "\033[30m" {
		t.Fail()
	}
}
