package wmic

import (
	"fmt"
	"log"
	"os/exec"
	"testing"
	"time"
)

type win32Service struct {
	Name        string
	DisplayName string
	StartMode   string
	StartName   string
	PathName    string
	State       string
}

type perfResult struct {
	IDProcess            int
	ElapsedTime          uint64
	PercentProcessorTime uint64
	ThreadCount          uint64
	WorkingSet           uint64
}

func TestCMDService(t *testing.T) {

	start := time.Now()
	cmd := exec.Command("wmic", "PATH", "Win32_Service", "WHERE", "(Status", "LIKE", "'%tm1sd%')", "GET", "Name,DisplayName,StartMode,StartName,PathName,State", "/format:csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s:\n%s\n", time.Since(start), string(out))

}

func TestCMDServiceAll(t *testing.T) {

	start := time.Now()
	cmd := exec.Command("wmic", "PATH", "Win32_Service", "GET", "Name,DisplayName,StartMode,StartName,PathName,State", "/format:csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s:\n%s\n", time.Since(start), string(out))

}

func TestCMDPerfAll(t *testing.T) {

	start := time.Now()
	cmd := exec.Command("wmic", "PATH", "Win32_PerfFormattedData_PerfProc_Process", "GET", "IDProcess,Name,ElapsedTime,PercentProcessorTime,ThreadCount,WorkingSet", "/format:csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s:\n%s\n", time.Since(start), string(out))

}

func TestCMDProcessor(t *testing.T) {

	start := time.Now()
	cmd := exec.Command("wmic", "PATH", "Win32_PerfFormattedData_PerfProcessor", "GET", "/format:csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s:\n%s\n", time.Since(start), string(out))

}

func TestCMDPerf(t *testing.T) {

	start := time.Now()
	cmd := exec.Command("wmic", "PATH", "Win32_PerfFormattedData_PerfProc_Process", "WHERE", "(", "IDProcess", "=", "6808", ")", "GET", "IDProcess,ElapsedTime,PercentProcessorTime,ThreadCount,WorkingSet", "/format:csv")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	fmt.Printf("%s:\n%s\n", time.Since(start), string(out))

}

func TestServiceAll(t *testing.T) {

	out := []*win32Service{}
	start := time.Now()
	err := QueryAll("Win32_Processor", &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}

func TestServiceColumns(t *testing.T) {

	out := []*win32Service{}
	start := time.Now()
	err := QueryColumns("Win32_Service", []string{"Name", "DisplayName", "StartMode", "StartName", "PathName", "State"}, &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}

func TestServiceColumnsWhere(t *testing.T) {

	out := []*win32Service{}
	start := time.Now()
	err := Query("Win32_Service", []string{"Name", "DisplayName", "StartMode", "StartName", "PathName", "State"}, "(PathName LIKE '%tm1sd%')", &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}

func TestServiceWhere(t *testing.T) {

	out := []*win32Service{}
	start := time.Now()
	err := QueryWhere("Win32_Service", "(PathName LIKE '%tm1sd%')", &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}

func TestProcess(t *testing.T) {

	out := []*perfResult{}
	start := time.Now()
	err := QueryColumns("Win32_PerfFormattedData_PerfProc_Process", []string{"IDProcess", "ElapsedTime", "PercentProcessorTime", "ThreadCount", "WorkingSet"}, &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}

func TestProcessWhere(t *testing.T) {

	out := []*perfResult{}
	start := time.Now()
	err := Query("Win32_PerfFormattedData_PerfProc_Process", []string{"IDProcess", "ElapsedTime", "PercentProcessorTime", "ThreadCount", "WorkingSet"}, "(IDProcess=15276 or IDProcess=1068 or IDProcess=4640)", &out)
	if err != nil {
		log.Fatalf("wmi query failed: %s", err)
	}
	fmt.Printf("%s:%v", time.Since(start), len(out))

}
