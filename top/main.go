package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"

	"github.com/casadebackend/goprobe/internal/checker"
	"github.com/casadebackend/goprobe/internal/sysinfo"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "tools":
		cmdTools()
	case "sysinfo":
		cmdSysInfo()
	case "version":
		fmt.Printf("goprobe v%s\n", version)
	default:
		color.Red("  comando desconhecido: %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	bold := color.New(color.Bold, color.FgCyan)
	bold.Printf("\ngoprobe v%s — checklist do ambiente de desenvolvimento\n\n", version)
	fmt.Println("Comandos:")
	fmt.Println("  tools           verifica ferramentas instaladas no sistema")
	fmt.Println("  sysinfo         exibe informações do sistema operacional")
	fmt.Println("  version         exibe a versão do goprobe")
	fmt.Println()
}

func cmdTools() {
	header := color.New(color.Bold, color.FgCyan)
	header.Println("\n  Ferramentas Instaladas")

	results := checker.CheckAll(checker.DefaultTools)
	for _, r := range results {
		if r.Found {
			color.Green("  ✓  %-10s  %s  (%v)", r.Tool, r.Version, r.Latency.Round(time.Millisecond))
		} else {
			color.Red("  ✗  %-10s  não encontrado", r.Tool)
		}
	}
	fmt.Println()
}

func cmdSysInfo() {
	header := color.New(color.Bold, color.FgCyan)
	header.Println("\n  Informações do Sistema")

	info := sysinfo.Get()
	fmt.Printf("  OS:       %s\n", info.OS)
	fmt.Printf("  Arch:     %s\n", info.Arch)
	fmt.Printf("  CPU:      %s\n", info.CPUInfo)
	fmt.Printf("  Memória:  %s\n", info.MemInfo)
	fmt.Println()
}
